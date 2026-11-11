package bot

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgtest"
	"github.com/gotd/td/tgtest/cluster"
	"github.com/gotd/td/transport"
)

const (
	testDC    = 2
	testBotID = 123456
	testToken = "123456:secret"
	testUser  = 42
)

// fakeServer is the state of a test Telegram server with one account.
type fakeServer struct {
	mux sync.Mutex
	// account is the authorized account, nil if not authorized.
	account *tg.User
	// imports counts auth.importBotAuthorization calls.
	imports int
	// importErrs are returned by the next auth.importBotAuthorization calls.
	importErrs []*tgerr.Error
	// pts is the server pts; messages[i] has pts i+1.
	messages []string
}

func (s *fakeServer) addMessage(text string) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.messages = append(s.messages, text)
}

func (s *fakeServer) stats() (imports int) {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.imports
}

func (s *fakeServer) state() tg.UpdatesState {
	return tg.UpdatesState{Pts: len(s.messages), Qts: 0, Date: 1, Seq: 0}
}

func (s *fakeServer) register(d *tgtest.Dispatcher) {
	d.HandleFunc(tg.HelpGetConfigRequestTypeID, func(server *tgtest.Server, req *tgtest.Request) error {
		return server.SendResult(req, &tg.Config{ThisDC: testDC})
	})
	d.HandleFunc(tg.UsersGetUsersRequestTypeID, func(server *tgtest.Server, req *tgtest.Request) error {
		s.mux.Lock()
		account := s.account
		s.mux.Unlock()
		if account == nil {
			return server.SendErr(req, tgerr.New(401, "AUTH_KEY_UNREGISTERED"))
		}
		return server.SendVector(req, account)
	})
	d.HandleFunc(tg.AuthImportBotAuthorizationRequestTypeID, func(server *tgtest.Server, req *tgtest.Request) error {
		var r tg.AuthImportBotAuthorizationRequest
		if err := r.Decode(req.Buf); err != nil {
			return err
		}
		s.mux.Lock()
		s.imports++
		if len(s.importErrs) > 0 {
			e := s.importErrs[0]
			s.importErrs = s.importErrs[1:]
			s.mux.Unlock()
			return server.SendErr(req, e)
		}
		tok, err := auth.ParseBotToken(r.BotAuthToken)
		if err != nil {
			s.mux.Unlock()
			return server.SendErr(req, tgerr.New(400, "ACCESS_TOKEN_INVALID"))
		}
		s.account = &tg.User{ID: tok.ID, Self: true, Bot: true, AccessHash: 1, FirstName: "bot"}
		s.account.SetFlags()
		account := s.account
		s.mux.Unlock()
		return server.SendResult(req, &tg.AuthAuthorization{User: account})
	})
	d.HandleFunc(tg.UpdatesGetStateRequestTypeID, func(server *tgtest.Server, req *tgtest.Request) error {
		s.mux.Lock()
		st := s.state()
		s.mux.Unlock()
		return server.SendResult(req, &st)
	})
	d.HandleFunc(tg.UpdatesGetDifferenceRequestTypeID, func(server *tgtest.Server, req *tgtest.Request) error {
		var r tg.UpdatesGetDifferenceRequest
		if err := r.Decode(req.Buf); err != nil {
			return err
		}
		s.mux.Lock()
		defer s.mux.Unlock()
		st := s.state()
		if r.Pts >= len(s.messages) {
			return server.SendResult(req, &tg.UpdatesDifferenceEmpty{Date: st.Date, Seq: st.Seq})
		}
		from := &tg.User{ID: testUser, AccessHash: 7, FirstName: "user"}
		from.SetFlags()
		diff := &tg.UpdatesDifference{
			Users: []tg.UserClass{from},
			State: st,
		}
		for i := r.Pts; i < len(s.messages); i++ {
			diff.NewMessages = append(diff.NewMessages, &tg.Message{
				ID:      i + 1,
				PeerID:  &tg.PeerUser{UserID: testUser},
				FromID:  &tg.PeerUser{UserID: testUser},
				Message: s.messages[i],
				Date:    1,
			})
		}
		return server.SendResult(req, diff)
	})
}

// env is a test cluster with a fakeServer.
type env struct {
	t       *testing.T
	cluster *cluster.Cluster
	server  *fakeServer
	ctx     context.Context
}

func newEnv(t *testing.T) *env {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	c := cluster.NewCluster(cluster.Options{
		Logger:   logzap.New(zaptest.NewLogger(t).Named("cluster")),
		Protocol: transport.Intermediate,
	})
	s := &fakeServer{}
	s.register(c.Dispatch(testDC, "dc"))

	g := tdsync.NewCancellableGroup(ctx)
	g.Go(c.Up)
	t.Cleanup(func() {
		g.Cancel()
		if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("cluster: %v", err)
		}
	})
	select {
	case <-c.Ready():
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	return &env{t: t, cluster: c, server: s, ctx: ctx}
}

func (e *env) options(dir string, h telegram.UpdateHandler) Options {
	return Options{
		Dir:     dir,
		Handler: h,
		Client: telegram.Options{
			PublicKeys: e.cluster.Keys(),
			Resolver:   e.cluster.Resolver(),
			DCList:     e.cluster.List(),
			DC:         testDC,
			Logger:     logzap.New(zaptest.NewLogger(e.t).Named("client")),
		},
	}
}

// messages collects texts of new messages.
type messages struct {
	mux   sync.Mutex
	texts []string
	got   chan struct{}
}

func newMessages() *messages {
	return &messages{got: make(chan struct{}, 100)}
}

func (m *messages) handler() telegram.UpdateHandler {
	d := tg.NewUpdateDispatcher()
	d.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
		msg, ok := u.Message.(*tg.Message)
		if !ok {
			return nil
		}
		m.mux.Lock()
		m.texts = append(m.texts, msg.Message)
		m.mux.Unlock()
		m.got <- struct{}{}
		return nil
	})
	return d
}

func (m *messages) wait(t *testing.T, ctx context.Context, n int) []string {
	t.Helper()
	for i := 0; i < n; i++ {
		select {
		case <-m.got:
		case <-ctx.Done():
			t.Fatalf("waiting for %d messages: %v", n, ctx.Err())
		}
	}
	m.mux.Lock()
	defer m.mux.Unlock()
	return append([]string(nil), m.texts...)
}

func TestBotRestartGetsMissedUpdates(t *testing.T) {
	e := newEnv(t)
	dir := t.TempDir()

	// First run: signs in, stores the session and the update state.
	b, err := Start(e.ctx, 1, "hash", testToken, e.options(dir, nil))
	require.NoError(t, err)
	require.Equal(t, int64(testBotID), b.Self().ID)
	require.Equal(t, int64(testBotID), b.ID())
	require.NoError(t, b.Close())
	require.Equal(t, 1, e.server.stats())

	// Messages sent while the bot is stopped.
	e.server.addMessage("one")
	e.server.addMessage("two")

	// Second run: the session is reused, the missed messages are delivered.
	msgs := newMessages()
	b, err = Start(e.ctx, 1, "hash", testToken, e.options(dir, msgs.handler()))
	require.NoError(t, err)
	require.Equal(t, []string{"one", "two"}, msgs.wait(t, e.ctx, 2))
	require.Equal(t, 1, e.server.stats(), "the stored session must be reused")

	// The peer manager got the entities of the difference.
	u, err := b.Peers().GetUser(e.ctx, &tg.InputUser{UserID: testUser, AccessHash: 7})
	require.NoError(t, err)
	require.Equal(t, "user", u.Raw().FirstName)

	select {
	case <-b.Done():
		t.Fatal("bot stopped")
	default:
	}
	require.NoError(t, b.Close())
	require.NoError(t, b.Wait(e.ctx))
	select {
	case <-b.Done():
	default:
		t.Fatal("Done must be closed after Close")
	}
}

func TestBotSessionMismatch(t *testing.T) {
	e := newEnv(t)
	dir := t.TempDir()

	b, err := Start(e.ctx, 1, "hash", testToken, e.options(dir, nil))
	require.NoError(t, err)
	require.NoError(t, b.Close())

	// Same session, another bot.
	_, err = Start(e.ctx, 1, "hash", "654321:other", e.options(dir, nil))
	require.ErrorIs(t, err, ErrSessionMismatch)
	require.Equal(t, 1, e.server.stats(), "no sign in into another account")
}

func TestBotLoginFloodWait(t *testing.T) {
	e := newEnv(t)
	e.server.importErrs = []*tgerr.Error{tgerr.New(420, "FLOOD_WAIT_1")}

	b, err := Start(e.ctx, 1, "hash", testToken, e.options("", nil))
	require.NoError(t, err)
	require.Equal(t, 2, e.server.stats())
	require.NoError(t, b.Close())
}

func TestBotLoginFloodWaitTooLong(t *testing.T) {
	e := newEnv(t)
	e.server.importErrs = []*tgerr.Error{tgerr.New(420, "FLOOD_WAIT_3600")}

	opt := e.options("", nil)
	opt.MaxLoginWait = time.Minute
	_, err := Start(e.ctx, 1, "hash", testToken, opt)
	d, ok := tgerr.AsFloodWait(err)
	require.True(t, ok, "got %v", err)
	require.Equal(t, time.Hour, d)
	require.Equal(t, 1, e.server.stats())
}

func TestBotInvalidToken(t *testing.T) {
	e := newEnv(t)
	e.server.importErrs = []*tgerr.Error{tgerr.New(400, "ACCESS_TOKEN_INVALID")}

	b, err := New(1, "hash", testToken, e.options("", nil))
	require.NoError(t, err)
	err = b.Start(e.ctx)
	require.True(t, tgerr.Is(err, "ACCESS_TOKEN_INVALID"), "got %v", err)
	require.Equal(t, 1, e.server.stats(), "not retried")

	// Stopped after a failed Start.
	select {
	case <-b.Done():
	default:
		t.Fatal("Done must be closed after a failed Start")
	}
	require.NoError(t, b.Close())
	require.ErrorIs(t, b.Start(e.ctx), telegram.ErrAlreadyStarted)
}

func TestNewErrors(t *testing.T) {
	_, err := New(1, "hash", "not a token", Options{})
	require.ErrorIs(t, err, auth.ErrInvalidBotToken)

	_, err = New(1, "hash", testToken, Options{Client: telegram.Options{
		UpdateHandler: telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error { return nil }),
	}})
	require.Error(t, err)

	// Stop before Start.
	b, err := New(1, "hash", testToken, Options{})
	require.NoError(t, err)
	b.Stop()
	require.ErrorIs(t, b.Start(context.Background()), telegram.ErrClientClosed)
	require.NoError(t, b.Close())
}
