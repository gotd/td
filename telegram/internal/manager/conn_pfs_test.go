package manager

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/tg"
)

type sessionCaptureHandler struct {
	cfg      tg.Config
	session  mtproto.Session
	sessions []mtproto.Session
	calls    int
}

func (h *sessionCaptureHandler) OnSession(cfg tg.Config, session mtproto.Session) error {
	h.cfg = cfg
	h.session = session
	h.sessions = append(h.sessions, session)
	h.calls++
	return nil
}

func (*sessionCaptureHandler) OnMessage(*bin.Buffer) error {
	return nil
}

func TestConnOnSessionDeferredUntilConfig(t *testing.T) {
	a := require.New(t)
	handler := &sessionCaptureHandler{}
	c := &Conn{
		log:         zap.NewNop(),
		handler:     handler,
		sessionInit: tdsync.NewReady(),
		gotConfig:   tdsync.NewReady(),
		dead:        tdsync.NewReady(),
	}
	s := mtproto.Session{ID: 123}

	// Quote (PFS): "Once auth.bindTempAuthKey has been executed successfully, the client can continue generating API calls as usual."
	// Link: https://core.telegram.org/api/pfs
	//
	// initConnection happens after bind, so OnSession may arrive before config.
	a.NoError(c.OnSession(s))
	a.Equal(0, handler.calls)

	c.mux.Lock()
	c.cfg = tg.Config{ThisDC: 2}
	c.mux.Unlock()
	c.gotConfig.Signal()

	a.NoError(c.flushPendingSession())
	a.Equal(1, handler.calls)
	a.Equal(2, handler.cfg.ThisDC)
	a.Equal(s, handler.session)
}

func TestConnOnSessionImmediateWithConfig(t *testing.T) {
	a := require.New(t)
	handler := &sessionCaptureHandler{}
	c := &Conn{
		log:         zap.NewNop(),
		handler:     handler,
		sessionInit: tdsync.NewReady(),
		gotConfig:   tdsync.NewReady(),
		dead:        tdsync.NewReady(),
	}
	c.mux.Lock()
	c.cfg = tg.Config{ThisDC: 4}
	c.mux.Unlock()
	c.gotConfig.Signal()

	s := mtproto.Session{ID: 999}
	a.NoError(c.OnSession(s))
	a.Equal(1, handler.calls)
	a.Equal(4, handler.cfg.ThisDC)
	a.Equal(s, handler.session)
}

func TestConnOnSessionDeferredQueue(t *testing.T) {
	a := require.New(t)
	handler := &sessionCaptureHandler{}
	c := &Conn{
		log:         zap.NewNop(),
		handler:     handler,
		sessionInit: tdsync.NewReady(),
		gotConfig:   tdsync.NewReady(),
		dead:        tdsync.NewReady(),
	}
	s1 := mtproto.Session{ID: 111}
	s2 := mtproto.Session{ID: 222}

	a.NoError(c.OnSession(s1))
	a.NoError(c.OnSession(s2))
	a.Equal(0, handler.calls)

	c.mux.Lock()
	c.cfg = tg.Config{ThisDC: 7}
	c.mux.Unlock()
	c.gotConfig.Signal()

	a.NoError(c.flushPendingSession())
	// Both events should be delivered, not overwritten by last one.
	a.Equal(2, handler.calls)
	a.Equal([]mtproto.Session{s1, s2}, handler.sessions)
}

func TestConnWaitSessionRequiresConfig(t *testing.T) {
	a := require.New(t)
	c := &Conn{
		sessionInit: tdsync.NewReady(),
		gotConfig:   tdsync.NewReady(),
		dead:        tdsync.NewReady(),
	}
	c.sessionInit.Signal()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	err := c.waitSession(ctx)
	a.Error(err)
	a.ErrorIs(err, context.DeadlineExceeded)

	c.gotConfig.Signal()
	// As soon as config is known, waitSession should unblock.
	a.NoError(c.waitSession(context.Background()))
}

func TestConnReadyUsesConfigSignal(t *testing.T) {
	a := require.New(t)
	c := &Conn{
		sessionInit: tdsync.NewReady(),
		gotConfig:   tdsync.NewReady(),
		dead:        tdsync.NewReady(),
	}
	c.sessionInit.Signal()

	select {
	case <-c.Ready():
		a.Fail("ready should not be signaled before config")
	default:
	}

	c.gotConfig.Signal()
	select {
	case <-c.Ready():
	case <-time.After(time.Second):
		a.Fail("ready should be signaled after config")
	}
}
