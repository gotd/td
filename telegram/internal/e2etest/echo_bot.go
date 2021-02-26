package e2etest

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// EchoBot is a simple echo message bot.
type EchoBot struct {
	suite *Suite

	logger *zap.Logger
	auth   chan<- *tg.User
}

// NewEchoBot creates new echo bot.
func NewEchoBot(suite *Suite, auth chan<- *tg.User) EchoBot {
	return EchoBot{
		suite:  suite,
		logger: suite.Log.Named("echobot"),
		auth:   auth,
	}
}

type users struct {
	users map[int]*tg.User
	lock  sync.RWMutex
}

func newUsers() *users {
	return &users{
		users: map[int]*tg.User{},
	}
}

func (m *users) empty() (r bool) {
	m.lock.RLock()
	r = len(m.users) < 1
	m.lock.RUnlock()
	return
}

func (m *users) add(list ...tg.UserClass) {
	m.lock.Lock()
	defer m.lock.Unlock()

	tg.UserClassArray(list).FillNotEmptyMap(m.users)
}

func (m *users) get(id int) (r *tg.User) {
	m.lock.RLock()
	r = m.users[id]
	m.lock.RUnlock()

	return
}

func (b EchoBot) login(ctx context.Context, client *telegram.Client) (*tg.User, error) {
	if err := b.suite.RetryAuthenticate(ctx, client); err != nil {
		return nil, xerrors.Errorf("authenticate: %w", err)
	}

	me, err := client.Self(ctx)
	if err != nil {
		return nil, xerrors.Errorf("get self: %w", err)
	}

	expectedUsername := "echobot" + strconv.Itoa(me.ID)
	raw := tg.NewClient(client)
	_, err = raw.AccountUpdateUsername(ctx, expectedUsername)
	if err != nil {
		var rpcErr *mtproto.Error
		if !errors.As(err, &rpcErr) || rpcErr.Message != "USERNAME_NOT_MODIFIED" {
			return nil, xerrors.Errorf("update username: %w", err)
		}
	}

	err = backoff.Retry(func() error {
		me, err = client.Self(ctx)
		if err != nil {
			return backoff.Permanent(xerrors.Errorf("get self: %w", err))
		}

		if me.Username != expectedUsername {
			return xerrors.Errorf("expected username %q, got %q", expectedUsername, me.Username)
		}

		return nil
	}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))
	if err != nil {
		return nil, err
	}

	return me, nil
}

func (b EchoBot) handler(client *telegram.Client) tg.NewMessageHandler {
	dialogsUsers := newUsers()

	return func(ctx tg.UpdateContext, update *tg.UpdateNewMessage) error {
		if filterMessage(update) {
			return nil
		}

		if m, ok := update.Message.(interface{ GetMessage() string }); ok {
			b.logger.Named("dispatcher").
				Info("Got new message update", zap.String("message", m.GetMessage()))
		}

		if dialogsUsers.empty() {
			raw := tg.NewClient(client)
			dialogs, err := raw.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
				Limit:      100,
				OffsetPeer: &tg.InputPeerEmpty{},
			})
			if err != nil {
				return xerrors.Errorf("get dialogs: %w", err)
			}

			if dlg, ok := dialogs.AsModified(); ok {
				dialogsUsers.add(dlg.GetUsers()...)
			}
		}

		switch m := update.Message.(type) {
		case *tg.Message:
			switch peer := m.PeerID.(type) {
			case *tg.PeerUser:
				user := ctx.Users[peer.UserID]
				if user == nil {
					user = dialogsUsers.get(peer.UserID)
				}

				b.logger.Info("Got message",
					zap.String("text", m.Message),
					zap.Int("user_id", user.ID),
					zap.String("user_first_name", user.FirstName),
					zap.String("username", user.Username),
				)

				randomID, err := client.RandInt64()
				if err != nil {
					return err
				}

				msg := &tg.MessagesSendMessageRequest{
					RandomID: randomID,
					Message:  m.Message,
					Peer: &tg.InputPeerUserFromMessage{
						Peer:   user.AsInputPeer(),
						UserID: peer.UserID,
						MsgID:  m.ID,
					},
				}

				if err := client.SendMessage(ctx, msg); err != nil {
					return xerrors.Errorf("send message: %w", err)
				}
				return nil
			}
		}

		return nil
	}
}

// Run setups and starts echo bot.
func (b EchoBot) Run(ctx context.Context) error {
	dispatcher := tg.NewUpdateDispatcher()
	client := b.suite.Client(b.logger, dispatcher)
	dispatcher.OnNewMessage(b.handler(client))

	return client.Run(ctx, func(ctx context.Context) error {
		defer close(b.auth)

		me, err := b.login(ctx, client)
		if err != nil {
			return xerrors.Errorf("login: %w", err)
		}

		b.logger.Info("Logged in",
			zap.String("user", me.Username),
			zap.Int("id", me.ID),
		)

		select {
		case b.auth <- me:
		case <-ctx.Done():
			return ctx.Err()
		}

		<-ctx.Done()
		return nil
	})
}
