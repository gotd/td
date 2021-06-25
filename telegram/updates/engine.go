package updates

import (
	"context"
	"sync"
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"

	"go.uber.org/atomic"
	"go.uber.org/zap"
)

const (
	idleTimeout    = time.Minute * 15
	fastgapTimeout = time.Millisecond * 500

	diffLimitUser = 100
	diffLimitBot  = 100000
)

// Engine deals with gaps.
type Engine struct {
	// Common state.
	pts         *sequenceBox
	qts         *sequenceBox
	seq         *sequenceBox
	date        int
	dateMux     sync.Mutex
	recovering  atomic.Bool
	idleTimeout *time.Timer

	// Channels state.
	channels map[int]*channelState
	chanMux  sync.Mutex

	// Channel access hashes.
	// Needed to perform updates.getChannelDifference.
	// Obtained lazily.
	channelHashes *hashStorage

	uchan      chan tg.UpdatesClass
	recoverGap chan struct{}

	shutdownMux sync.Mutex
	closed      bool
	done        chan struct{}

	diffLim int
	selfID  int
	storage Storage
	raw     RawClient
	handler Handler
	forget  bool
	log     *zap.Logger

	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup
}

// New creates new engine.
func New(cfg Config) *Engine {
	cfg.setDefaults()

	ctx, cancel := context.WithCancel(context.Background())
	e := &Engine{
		idleTimeout: time.NewTimer(idleTimeout),
		channels:    map[int]*channelState{},
		channelHashes: &hashStorage{
			hasher: cfg.AccessHasher,
			log:    cfg.Logger.Named("hasher"),
		},

		uchan:      make(chan tg.UpdatesClass, 10),
		recoverGap: make(chan struct{}),
		done:       make(chan struct{}),

		diffLim: diffLimitUser,
		selfID:  cfg.SelfID,
		storage: cfg.Storage,
		raw:     cfg.RawClient,
		handler: cfg.Handler,
		forget:  cfg.Forget,
		log:     cfg.Logger,

		ctx:    ctx,
		cancel: cancel,
	}

	if cfg.IsBot {
		e.diffLim = diffLimitBot
	}

	return e
}

func (e *Engine) handleUpdates(u tg.UpdatesClass) error {
	_ = e.idleTimeout.Reset(idleTimeout)
	switch u := u.(type) {
	case *tg.Updates:
		e.saveChannelHashes("Updates", u.Chats)
		return e.handleSeq(&tg.UpdatesCombined{
			Updates:  u.Updates,
			Users:    u.Users,
			Chats:    u.Chats,
			Date:     u.Date,
			Seq:      u.Seq,
			SeqStart: u.Seq,
		})

	case *tg.UpdatesCombined:
		e.saveChannelHashes("UpdatesCombined", u.Chats)
		return e.handleSeq(u)

	case *tg.UpdateShort:
		return e.handleUpdates(&tg.UpdatesCombined{
			Updates: []tg.UpdateClass{u.Update},
			Date:    u.Date,
		})

	case *tg.UpdateShortMessage:
		return e.handleUpdates(e.convertShortMessage(u))

	case *tg.UpdateShortChatMessage:
		return e.handleUpdates(e.convertShortChatMessage(u))

	case *tg.UpdateShortSentMessage:
		return e.handleUpdates(e.convertShortSentMessage(u))

	case *tg.UpdatesTooLong:
		e.recoverGap <- struct{}{}
		return nil

	default:
		return xerrors.Errorf("unexpected tg.UpdatesClass type: %T", u)
	}
}

func (e *Engine) handleSeq(u *tg.UpdatesCombined) error {
	if err := validateSeq(u.Seq, u.SeqStart); err != nil {
		return xerrors.Errorf("validate seq: %w", err)
	}

	// Special case.
	if u.Seq == 0 {
		ptsChanged, err := e.applyCombined(u)
		if err != nil {
			return err
		}

		if ptsChanged {
			e.recoverGap <- struct{}{}
		}
		return nil
	}

	return e.seq.Handle(update{
		Value: u,
		State: u.Seq,
		Count: u.Seq - u.SeqStart + 1,
	})
}

func (e *Engine) handlePts(pts, ptsCount int, u tg.UpdateClass, ents *Entities) error {
	if err := validatePts(pts, ptsCount); err != nil {
		e.log.Warn("Pts validation failed", zap.Error(err))
		return nil
	}

	return e.pts.Handle(update{
		Value: u,
		State: pts,
		Count: ptsCount,
		Ents:  ents,
	})
}

func (e *Engine) handleQts(qts int, u tg.UpdateClass, ents *Entities) error {
	if err := validateQts(qts); err != nil {
		e.log.Warn("Qts validation failed", zap.Error(err))
		return nil
	}

	return e.qts.Handle(update{
		Value: u,
		State: qts,
		Count: 1,
		Ents:  ents,
	})
}

func (e *Engine) handleChannel(channelID, date, pts, ptsCount int, u tg.UpdateClass, ents *Entities) error {
	if err := validatePts(pts, ptsCount); err != nil {
		e.log.Warn("Pts validation failed", zap.Error(err))
		return nil
	}

	e.chanMux.Lock()
	state, ok := e.channels[channelID]
	if !ok {
		state = e.createChannelState(channelID, pts-ptsCount)
		e.channels[channelID] = state
		if e.restoreHash(channelID, date) {
			state.recoverGap <- struct{}{}
		}
	}
	e.chanMux.Unlock()

	_ = state.idleTimeout.Reset(idleTimeout)
	return state.pts.Handle(update{
		Value: u,
		State: pts,
		Count: ptsCount,
		Ents:  ents,
	})
}

func (e *Engine) handleChannelTooLong(date int, long *tg.UpdateChannelTooLong) {
	e.chanMux.Lock()
	defer e.chanMux.Unlock()

	log := e.log.With(zap.Int("channel_id", long.ChannelID))

	state, ok := e.channels[long.ChannelID]
	if !ok {
		pts, havePts := long.GetPts()
		if !havePts {
			log.Debug("Got UpdateChannelTooLong without pts field")
			return
		}

		state = e.createChannelState(long.ChannelID, pts)
		e.channels[long.ChannelID] = state
		if e.restoreHash(long.ChannelID, date) {
			state.recoverGap <- struct{}{}
		}
		return
	}

	state.recoverGap <- struct{}{}
}
