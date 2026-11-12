package user

import (
	"context"
	"path/filepath"
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
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/gotd/td/tgtest"
	"github.com/gotd/td/tgtest/cluster"
	"github.com/gotd/td/transport"
)

const (
	testDC      = 2
	testUserID  = 1000
	testFriend  = 42
	testChannel = 500
	testPhone   = "+15550001234"
	testCode    = "12345"
)

// fakeServer is the state of a test Telegram server with one account.
type fakeServer struct {
	mux sync.Mutex
	// account is the authorized account, nil if not authorized.
	account *tg.User
	// sendCodeErrs are returned by the next auth.sendCode calls.
	sendCodeErrs []*tgerr.Error
	// calls counts requests by method name.
	calls map[string]int
	// messages: messages[i] has pts i+1.
	messages []string
	// qrAccepted is set when the QR token is accepted.
	qrAccepted bool
	// session is the session of the last request.
	session tgtest.Session
	srv     *tgtest.Server
	// channelDiffs is signaled on updates.getChannelDifference.
	channelDiffs chan struct{}
}

func newFakeServer() *fakeServer {
	return &fakeServer{calls: map[string]int{}, channelDiffs: make(chan struct{}, 10)}
}

func (s *fakeServer) count(method string) int {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.calls[method]
}

func (s *fakeServer) authCalls() int {
	return s.count("auth.sendCode") + s.count("auth.signIn") + s.count("auth.exportLoginToken")
}

func (s *fakeServer) addMessage(text string) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.messages = append(s.messages, text)
}

func (s *fakeServer) setAccount(u *tg.User) {
	s.mux.Lock()
	defer s.mux.Unlock()
	u.Self = true
	u.SetFlags()
	s.account = u
}

func (s *fakeServer) authorize() *tg.User {
	// Called with mux held.
	s.account = &tg.User{ID: testUserID, Self: true, AccessHash: 1, FirstName: "me"}
	s.account.SetFlags()
	return s.account
}

func (s *fakeServer) state() tg.UpdatesState {
	return tg.UpdatesState{Pts: len(s.messages), Date: 1}
}

func (s *fakeServer) register(d *tgtest.Dispatcher) {
	handle := func(id uint32, name string, f func(server *tgtest.Server, req *tgtest.Request) error) {
		d.HandleFunc(id, func(server *tgtest.Server, req *tgtest.Request) error {
			s.mux.Lock()
			s.calls[name]++
			s.srv = server
			s.session = req.Session
			s.mux.Unlock()
			return f(server, req)
		})
	}
	handle(tg.HelpGetConfigRequestTypeID, "help.getConfig", func(server *tgtest.Server, req *tgtest.Request) error {
		return server.SendResult(req, &tg.Config{ThisDC: testDC})
	})
	handle(tg.UsersGetUsersRequestTypeID, "users.getUsers", func(server *tgtest.Server, req *tgtest.Request) error {
		s.mux.Lock()
		account := s.account
		s.mux.Unlock()
		if account == nil {
			return server.SendErr(req, tgerr.New(401, "AUTH_KEY_UNREGISTERED"))
		}
		return server.SendVector(req, account)
	})
	handle(tg.AuthSendCodeRequestTypeID, "auth.sendCode", func(server *tgtest.Server, req *tgtest.Request) error {
		s.mux.Lock()
		if len(s.sendCodeErrs) > 0 {
			e := s.sendCodeErrs[0]
			s.sendCodeErrs = s.sendCodeErrs[1:]
			s.mux.Unlock()
			return server.SendErr(req, e)
		}
		s.mux.Unlock()
		return server.SendResult(req, &tg.AuthSentCode{
			Type:          &tg.AuthSentCodeTypeApp{Length: len(testCode)},
			PhoneCodeHash: "hash",
		})
	})
	handle(tg.AuthSignInRequestTypeID, "auth.signIn", func(server *tgtest.Server, req *tgtest.Request) error {
		var r tg.AuthSignInRequest
		if err := r.Decode(req.Buf); err != nil {
			return err
		}
		if code, _ := r.GetPhoneCode(); code != testCode {
			return server.SendErr(req, tgerr.New(400, "PHONE_CODE_INVALID"))
		}
		s.mux.Lock()
		account := s.authorize()
		s.mux.Unlock()
		return server.SendResult(req, &tg.AuthAuthorization{User: account})
	})
	handle(tg.AuthExportLoginTokenRequestTypeID, "auth.exportLoginToken", func(server *tgtest.Server, req *tgtest.Request) error {
		s.mux.Lock()
		s.session = req.Session
		if s.qrAccepted {
			account := s.authorize()
			s.mux.Unlock()
			return server.SendResult(req, &tg.AuthLoginTokenSuccess{
				Authorization: &tg.AuthAuthorization{User: account},
			})
		}
		s.mux.Unlock()
		return server.SendResult(req, &tg.AuthLoginToken{
			Expires: int(time.Now().Add(time.Minute).Unix()),
			Token:   []byte("token"),
		})
	})
	handle(tg.UpdatesGetStateRequestTypeID, "updates.getState", func(server *tgtest.Server, req *tgtest.Request) error {
		s.mux.Lock()
		s.session = req.Session
		st := s.state()
		s.mux.Unlock()
		return server.SendResult(req, &st)
	})
	handle(tg.UpdatesGetDifferenceRequestTypeID, "updates.getDifference", func(server *tgtest.Server, req *tgtest.Request) error {
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
		from := &tg.User{ID: testFriend, AccessHash: 7, FirstName: "friend"}
		from.SetFlags()
		channel := &tg.Channel{ID: testChannel, AccessHash: 99, Title: "channel", Broadcast: true, Photo: &tg.ChatPhotoEmpty{}}
		channel.SetFlags()
		diff := &tg.UpdatesDifference{
			Users: []tg.UserClass{from},
			Chats: []tg.ChatClass{channel},
			State: st,
		}
		for i := r.Pts; i < len(s.messages); i++ {
			diff.NewMessages = append(diff.NewMessages, &tg.Message{
				ID:      i + 1,
				PeerID:  &tg.PeerUser{UserID: testFriend},
				FromID:  &tg.PeerUser{UserID: testFriend},
				Message: s.messages[i],
				Date:    1,
			})
		}
		return server.SendResult(req, diff)
	})
	handle(tg.UpdatesGetChannelDifferenceRequestTypeID, "updates.getChannelDifference", func(server *tgtest.Server, req *tgtest.Request) error {
		select {
		case s.channelDiffs <- struct{}{}:
		default:
		}
		return server.SendResult(req, &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: 1})
	})
}

// pushUpdates sends updates to the last session that called getState.
func (s *fakeServer) pushUpdates(ctx context.Context, updates ...tg.UpdateClass) error {
	s.mux.Lock()
	srv, session := s.srv, s.session
	s.mux.Unlock()
	return srv.SendUpdates(ctx, session, updates...)
}

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
	s := newFakeServer()
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

func (e *env) options(dir string, h telegram.UpdateHandler, a Authenticator) Options {
	return Options{
		Dir:     dir,
		Handler: h,
		Auth:    a,
		Client: telegram.Options{
			PublicKeys: e.cluster.Keys(),
			Resolver:   e.cluster.Resolver(),
			DCList:     e.cluster.List(),
			DC:         testDC,
			Logger:     logzap.New(zaptest.NewLogger(e.t).Named("client")),
		},
	}
}

func phoneAuthenticator() Authenticator {
	code := auth.CodeAuthenticatorFunc(func(context.Context, *tg.AuthSentCode) (string, error) {
		return testCode, nil
	})
	return Phone(auth.Constant(testPhone, "", code), auth.SendCodeOptions{})
}

// messages collects texts of new messages.
type messages struct {
	mux   sync.Mutex
	texts []string
	got   chan struct{}
}

func newMessages() *messages { return &messages{got: make(chan struct{}, 100)} }

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

func TestUserRestartGetsMissedUpdates(t *testing.T) {
	e := newEnv(t)
	dir := t.TempDir()

	// First run signs in by phone and stores the session and the state.
	u, err := Start(e.ctx, 1, "hash", e.options(dir, nil, phoneAuthenticator()))
	require.NoError(t, err)
	require.Equal(t, int64(testUserID), u.ID())
	require.NoError(t, u.Close())
	require.Equal(t, 1, e.server.count("auth.signIn"))

	e.server.addMessage("one")
	e.server.addMessage("two")

	// Second run: no authenticator needed, the session is reused.
	msgs := newMessages()
	u, err = Start(e.ctx, 1, "hash", e.options(dir, msgs.handler(), nil))
	require.NoError(t, err)
	require.Equal(t, []string{"one", "two"}, msgs.wait(t, e.ctx, 2))
	require.Equal(t, 1, e.server.count("auth.sendCode"), "no new sign in")
	require.Equal(t, 1, e.server.count("auth.signIn"), "no new sign in")

	// The peer manager got the entities of the difference.
	p, err := u.Peers().GetUser(e.ctx, &tg.InputUser{UserID: testFriend, AccessHash: 7})
	require.NoError(t, err)
	require.Equal(t, "friend", p.Raw().FirstName)
	require.NoError(t, u.Close())
	require.NoError(t, u.Wait(e.ctx))
}

func TestUserAccessHashesPersisted(t *testing.T) {
	e := newEnv(t)
	dir := t.TempDir()
	u, err := Start(e.ctx, 1, "hash", e.options(dir, nil, phoneAuthenticator()))
	require.NoError(t, err)
	require.NoError(t, u.Close())

	// A missed message: the difference carries the channel and the user.
	e.server.addMessage("one")
	u, err = Start(e.ctx, 1, "hash", e.options(dir, nil, nil))
	require.NoError(t, err)
	// Wait for the difference to be applied.
	require.Eventually(t, func() bool {
		h, ok, err := updates.NewFileStorage(filepath.Join(dir, UpdatesFile)).
			GetChannelAccessHash(e.ctx, testUserID, testChannel)
		return err == nil && ok && h == 99
	}, 10*time.Second, 20*time.Millisecond)
	require.NoError(t, u.Close())

	fs := updates.NewFileStorage(filepath.Join(dir, UpdatesFile))
	h, ok, err := fs.GetUserAccessHash(e.ctx, testUserID, testFriend)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, int64(7), h)

	// After a restart the first update of the known channel needs no
	// getDifference to learn its access hash.
	diffs := e.server.count("updates.getDifference")
	u, err = Start(e.ctx, 1, "hash", e.options(dir, nil, nil))
	require.NoError(t, err)
	// The startup difference of this run, in the background.
	require.Eventually(t, func() bool {
		return e.server.count("updates.getDifference") == diffs+1
	}, 10*time.Second, 10*time.Millisecond)
	before := e.server.count("updates.getDifference")
	require.NoError(t, e.server.pushUpdates(e.ctx, &tg.UpdateNewChannelMessage{
		Message: &tg.Message{
			ID:     1,
			PeerID: &tg.PeerChannel{ChannelID: testChannel},
			Date:   1,
		},
		Pts:      2,
		PtsCount: 1,
	}))
	select {
	case <-e.server.channelDiffs:
	case <-e.ctx.Done():
		t.Fatal("no getChannelDifference")
	}
	require.Equal(t, before, e.server.count("updates.getDifference"))
	require.NoError(t, u.Close())
}

func TestUserNotAuthorized(t *testing.T) {
	e := newEnv(t)
	u, err := New(1, "hash", e.options("", nil, nil))
	require.NoError(t, err)
	require.ErrorIs(t, u.Start(e.ctx), ErrNotAuthorized)
	require.Zero(t, e.server.authCalls(), "nothing sent to sign in")
	require.NoError(t, u.Close())
}

func TestUserBotSession(t *testing.T) {
	e := newEnv(t)
	e.server.setAccount(&tg.User{ID: 123, Bot: true, AccessHash: 1})
	_, err := Start(e.ctx, 1, "hash", e.options("", nil, phoneAuthenticator()))
	require.ErrorIs(t, err, ErrSessionMismatch)
	require.Zero(t, e.server.authCalls())
}

func TestUserMismatch(t *testing.T) {
	e := newEnv(t)
	e.server.setAccount(&tg.User{ID: testUserID, AccessHash: 1})
	opt := e.options("", nil, phoneAuthenticator())
	opt.UserID = testUserID + 1
	_, err := Start(e.ctx, 1, "hash", opt)
	require.ErrorIs(t, err, ErrSessionMismatch)
	require.Zero(t, e.server.authCalls())

	opt.UserID = testUserID
	u, err := Start(e.ctx, 1, "hash", opt)
	require.NoError(t, err)
	require.NoError(t, u.Close())
}

func TestUserPhoneFloodWait(t *testing.T) {
	e := newEnv(t)
	e.server.sendCodeErrs = []*tgerr.Error{tgerr.New(420, "FLOOD_WAIT_1")}
	u, err := Start(e.ctx, 1, "hash", e.options("", nil, phoneAuthenticator()))
	require.NoError(t, err)
	require.Equal(t, 2, e.server.count("auth.sendCode"))
	require.Equal(t, 1, e.server.count("auth.signIn"))
	require.NoError(t, u.Close())
}

func TestUserPhoneFloodWaitTooLong(t *testing.T) {
	e := newEnv(t)
	e.server.sendCodeErrs = []*tgerr.Error{tgerr.New(420, "FLOOD_WAIT_3600")}
	opt := e.options("", nil, phoneAuthenticator())
	opt.MaxLoginWait = time.Minute
	_, err := Start(e.ctx, 1, "hash", opt)
	d, ok := tgerr.AsFloodWait(err)
	require.True(t, ok, "got %v", err)
	require.Equal(t, time.Hour, d)
	require.Equal(t, 1, e.server.count("auth.sendCode"))
}

func TestUserQR(t *testing.T) {
	e := newEnv(t)
	var shown int
	a := QR(QROptions{Show: func(ctx context.Context, token qrlogin.Token) error {
		shown++
		// Accepted on another device: the server sends updateLoginToken.
		e.server.mux.Lock()
		e.server.qrAccepted = true
		e.server.mux.Unlock()
		return e.server.pushUpdates(ctx, &tg.UpdateLoginToken{})
	}})
	u, err := Start(e.ctx, 1, "hash", e.options("", nil, a))
	require.NoError(t, err)
	require.Equal(t, int64(testUserID), u.ID())
	require.Equal(t, 1, shown)
	require.Equal(t, 2, e.server.count("auth.exportLoginToken"))
	require.NoError(t, u.Close())
}

func TestNewErrors(t *testing.T) {
	_, err := New(1, "hash", Options{Client: telegram.Options{
		UpdateHandler: telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error { return nil }),
	}})
	require.Error(t, err)

	// Stop before Start.
	u, err := New(1, "hash", Options{})
	require.NoError(t, err)
	u.Stop()
	require.ErrorIs(t, u.Start(context.Background()), telegram.ErrClientClosed)
	require.NoError(t, u.Close())
	require.Zero(t, u.ID())
}
