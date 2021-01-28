// Package pool contains implementation of MTProto connection pool.
package pool

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.uber.org/atomic"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/session"
	"github.com/gotd/td/tg"
)

// Pool represents multi-DC connection pool, literally pool of connection pools.
type Pool struct {
	// MTProto connection options. Can be mutated in restoreConnection.
	opts    mtproto.Options
	optsMux sync.Mutex

	// InitConnection parameters.
	appID int // immutable
	// Telegram device information.
	device DeviceConfig // immutable

	// Session storage.
	storage SessionStorage // immutable,nilable

	// Handler passed by client.
	handler ConnHandler // immutable

	// Wrappers for external world, like logs or PRNG.
	log *zap.Logger // immutable

	// Pool context. Will be canceled by Run on exit.
	ctx    context.Context    // immutable
	cancel context.CancelFunc // immutable

	// DCs supervisor.
	grp *tdsync.Supervisor

	// Primary DC address and ID.
	primary *dcInfo
	// DC connections.
	dcs    map[dcID]*DC
	dcsMux sync.Mutex

	// Limit of connections.
	max int64 // immutable

	// DC migration per request limit.
	migrationLimit int // immutable

	// Requests wait group.
	ongoing sync.WaitGroup

	// Current Telegram config.
	cfg *config

	closed atomic.Bool
}

// NewPool creates new uninitialized Pool.
func NewPool(appID int, handler ConnHandler, opts Options) *Pool {
	ctx, cancel := context.WithCancel(context.Background())

	opts.setDefaults()
	return &Pool{
		opts:    opts.MTProto,
		appID:   appID,
		device:  opts.Device,
		storage: opts.SessionStorage,
		handler: handler,
		log:     opts.Logger,
		ctx:     ctx,
		cancel:  cancel,
		grp:     tdsync.NewSupervisor(ctx),
		primary: &dcInfo{
			primaryDC:   dcID(opts.ID),
			primaryAddr: opts.Addr,
		},
		dcs:            map[dcID]*DC{},
		max:            opts.MaxOpenConnections,
		migrationLimit: opts.MigrationLimit,
		cfg:            newConfig(),
	}
}

// OnSession implements ConnHandler.
func (c *Pool) OnSession(addr string, cfg tg.Config, s mtproto.Session) error {
	_, primaryAddr := c.primary.Load()
	if primaryAddr == addr {
		if err := c.saveSession(addr, cfg, s); err != nil {
			return xerrors.Errorf("save: %w", err)
		}
	}

	c.cfg.Store(cfg)
	return nil
}

// OnMessage implements ConnHandler.
func (c *Pool) OnMessage(b *bin.Buffer) error {
	return c.handler.OnMessage(b)
}

func (c *Pool) restoreConnection(ctx context.Context) error {
	if c.storage == nil {
		return nil
	}
	data, err := c.storage.Load(ctx)
	if errors.Is(err, session.ErrNotFound) {
		return nil
	}
	if err != nil {
		return xerrors.Errorf("load: %w", err)
	}

	// Restoring persisted auth key.
	var key crypto.AuthKey
	copy(key.Value[:], data.AuthKey)
	copy(key.ID[:], data.AuthKeyID)

	if key.Value.ID() != key.ID {
		return xerrors.New("corrupted key")
	}

	// Re-initializing connection from persisted state.
	c.log.Info("Connection restored from state",
		zap.String("addr", data.Addr),
		zap.String("key_id", fmt.Sprintf("%x", data.AuthKeyID)),
	)

	c.optsMux.Lock()
	c.opts.Key = key
	c.opts.Salt = data.Salt
	c.optsMux.Unlock()
	c.primary.Store(dcID(data.DC), data.Addr)

	return nil
}

func (c *Pool) findAddress(ctx context.Context, id dcID) (string, error) {
	if primaryID, addr := c.primary.Load(); id == primaryID {
		return addr, nil
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-c.ctx.Done():
		return "", c.ctx.Err()
	case <-c.cfg.Ready(): // Await config receive.
	}

	addr, ok := c.cfg.FindAddress(int(id))
	if ok {
		return addr, nil
	}

	return "", xerrors.Errorf("address for DC %d not found", id)
}

func (c *Pool) updatePrimaryDC(id dcID) bool {
	addr, ok := c.cfg.FindAddress(int(id))
	if !ok {
		return false
	}

	c.primary.Store(id, addr)
	c.log.Info("New primary DC", zap.Int("dc_id", int(id)), zap.String("addr", addr))
	return true
}

func (c *Pool) deleteDC(ctx context.Context, id dcID) error {
	c.dcsMux.Lock()
	dc, ok := c.dcs[id]
	delete(c.dcs, id)
	c.dcsMux.Unlock()

	if !ok {
		return nil
	}

	return dc.Close(ctx)
}

func (c *Pool) acquireDC(ctx context.Context, id dcID) (*DC, error) {
	c.dcsMux.Lock()
	defer c.dcsMux.Unlock()

	dc, ok := c.dcs[id]
	if ok {
		return dc, nil
	}

	// If DC pool does not exist, so we create new DC pool.
	dc, err := c.createDC(ctx, id)
	if err != nil {
		return nil, err
	}

	// Add DC to DC map.
	c.dcs[id] = dc

	// Setup close handler.
	c.grp.Go(func(grpCtx context.Context) error {
		defer func() {
			_ = c.deleteDC(c.ctx, id)
		}()
		return dc.Run(grpCtx)
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	case <-dc.Ready():
	}

	return dc, nil
}

func (c *Pool) createDC(ctx context.Context, id dcID) (*DC, error) {
	addr, err := c.findAddress(ctx, id)
	if err != nil {
		return nil, err
	}

	c.optsMux.Lock()
	opts := c.opts
	c.optsMux.Unlock()

	dc := NewDC(id, addr, c, DCOptions{
		AppID:              c.appID,
		Device:             c.device,
		Logger:             c.log.Named("dc").With(zap.Int("dc_id", int(id))),
		MTProto:            opts,
		MaxOpenConnections: c.max,
	})
	return dc, nil
}

var errPoolIsClosed = xerrors.New("pool is closed")

func (c *Pool) invokeDC(ctx context.Context, id dcID, input bin.Encoder, output bin.Decoder) (err error) {
	dc, err := c.acquireDC(ctx, id)
	if err != nil {
		return xerrors.Errorf("acquire DC: %w", err)
	}

	return dc.InvokeRaw(ctx, input, output)
}

// InvokeRaw sends MTProto request using one of pool connection to primary DC.
// Retries request if got migration error.
func (c *Pool) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	if c.closed.Load() {
		return errPoolIsClosed
	}

	c.ongoing.Add(1)
	defer c.ongoing.Done()

	dc, _ := c.primary.Load()
	migration := 0
	for {
		err := c.invokeDC(ctx, dc, input, output)

		var rpcErr *mtproto.Error
		if errors.As(err, &rpcErr) && (rpcErr.Code == 303) {
			if migration == c.migrationLimit {
				return xerrors.Errorf(
					"migration limit (%d) exceed, DC: %d",
					c.migrationLimit, rpcErr.Argument,
				)
			}

			if rpcErr.Type == "STATS_MIGRATE" || rpcErr.Type == "FILE_MIGRATE" {
				c.updatePrimaryDC(dc)
			}

			c.log.Debug("Got migrate error: Starting migration to another dc",
				zap.String("error", rpcErr.Type),
				zap.Int("dc", rpcErr.Argument),
				zap.Int("migration", migration),
			)
			dc = dcID(rpcErr.Argument)

			migration++
			continue // Retry with given DC.
		}

		return err
	}
}

// Run initialize DC pool.
func (c *Pool) Run(ctx context.Context) error {
	if c.closed.Load() {
		return errPoolIsClosed
	}
	c.log.Debug("Starting connection pool")

	if err := c.restoreConnection(ctx); err != nil {
		return xerrors.Errorf("restore connection: %w", err)
	}

	id, _ := c.primary.Load()
	_, err := c.acquireDC(ctx, id)
	if err != nil {
		return xerrors.Errorf("acquire DC: %w", err)
	}

	<-ctx.Done()
	return c.close(ctx)
}

func (c *Pool) close(closeCtx context.Context) error {
	if c.closed.Swap(true) {
		return xerrors.New("Pool already closed")
	}
	c.log.Debug("Closing pool")

	// Cancel all DC tasks to start Close in DC.
	c.grp.Cancel()

	closed, cancel := context.WithCancel(closeCtx)
	go func() {
		c.ongoing.Wait()
		cancel()
	}()

	<-closed.Done()
	// Cancels all DCs, connections and pending requests.
	c.cancel()
	return c.grp.Wait()
}
