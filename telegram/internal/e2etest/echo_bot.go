package e2etest

import (
	"context"
	"strconv"
	"sync"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
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
		logger: suite.logger.Named("echobot"),
		auth:   auth,
	}
}

type users struct {
	users map[int64]*tg.User
	lock  sync.RWMutex
}

func newUsers() *users {
	return &users{
		users: map[int64]*tg.User{},
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

func (m *users) get(id int64) (r *tg.User) {
	m.lock.RLock()
	r = m.users[id]
	m.lock.RUnlock()

	return
}

func (b EchoBot) login(ctx context.Context, client *telegram.Client) (*tg.User, error) {
	if err := b.suite.RetryAuthenticate(ctx, client.Auth()); err != nil {
		return nil, errors.Wrap(err, "authenticate")
	}

	var me *tg.User
	if err := retry(ctx, func() (err error) {
		me, err = client.Self(ctx)
		return err
	}); err != nil {
		return nil, err
	}

	expectedUsername := "echobot" + strconv.FormatInt(me.ID, 10)
	raw := tg.NewClient(retryInvoker{prev: client})
	_, err := raw.AccountUpdateUsername(ctx, expectedUsername)
	if err != nil {
		if !tgerr.Is(err, tg.ErrUsernameNotModified) {
			return nil, errors.Wrap(err, "update username")
		}
	}
	me, err = retryResult(ctx, func() (*tg.User, error) {
		return client.Self(ctx)
	})
	if me.Username != expectedUsername {
		return nil, errors.Errorf("expected username %q, got %q", expectedUsername, me.Username)
	}

	return me, nil
}

func (b EchoBot) handler(client *telegram.Client) tg.NewMessageHandler {
	dialogsUsers := newUsers()

	raw := tg.NewClient(client)
	sender := message.NewSender(raw)
	return func(ctx context.Context, entities tg.Entities, update *tg.UpdateNewMessage) error {
		if filterMessage(update) {
			return nil
		}

		if m, ok := update.Message.(interface{ GetMessage() string }); ok {
			b.logger.Named("dispatcher").
				Info("Got new message update", zap.String("message", m.GetMessage()))
		}

		if dialogsUsers.empty() {
			dialogs, err := retryResult(ctx, func() (tg.MessagesDialogsClass, error) {
				dialogs, err := raw.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
					Limit:      100,
					OffsetPeer: &tg.InputPeerEmpty{},
				})
				if err != nil {
					return nil, errors.Wrap(err, "get dialogs")
				}
				return dialogs, nil
			})
			if err != nil {
				return errors.Wrap(err, "get dialogs")
			}
			if dlg, ok := dialogs.AsModified(); ok {
				dialogsUsers.add(dlg.GetUsers()...)
			}
		}

		switch m := update.Message.(type) {
		case *tg.Message:
			switch peer := m.PeerID.(type) {
			case *tg.PeerUser:
				user := entities.Users[peer.UserID]
				if user == nil {
					user = dialogsUsers.get(peer.UserID)
				}

				b.logger.Info("Got message",
					zap.String("text", m.Message),
					zap.Int64("user_id", user.ID),
					zap.String("user_first_name", user.FirstName),
					zap.String("username", user.Username),
				)

				if err := retry(ctx, func() error {
					_, err := sender.To(user.AsInputPeer()).Text(ctx, m.Message)
					return err
				}); err != nil {
					return errors.Wrap(err, "send message")
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
			return errors.Wrap(err, "login")
		}

		b.logger.Info("Logged in",
			zap.String("user", me.Username),
			zap.Int64("id", me.ID),
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
