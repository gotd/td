package e2etest

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram/message"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgerr"
)

// User is a simple user bot.
type User struct {
	suite    *Suite
	text     []string
	username string

	logger  *zap.Logger
	message chan string
}

// NewUser creates new User bot.
func NewUser(suite *Suite, text []string, username string) User {
	return User{
		suite:    suite,
		text:     text,
		username: username,
		logger:   suite.logger.Named("terentyev"),
		message:  make(chan string, 1),
	}
}

func (u User) messageHandler(ctx context.Context, entities tg.Entities, update *tg.UpdateNewMessage) error {
	if filterMessage(update) {
		return nil
	}

	if m, ok := update.Message.(interface{ GetMessage() string }); ok {
		u.logger.Named("dispatcher").
			Info("Got new message update", zap.String("message", m.GetMessage()))
	}

	msg, ok := update.Message.(*tg.Message)
	if !ok {
		return xerrors.Errorf("unexpected type %T", update.Message)
	}

	select {
	case u.message <- msg.Message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Run setups and starts user bot.
func (u User) Run(ctx context.Context) error {
	dispatcher := tg.NewUpdateDispatcher()
	dispatcher.OnNewMessage(u.messageHandler)
	client := u.suite.Client(u.logger, dispatcher)
	sender := message.NewSender(tg.NewClient(waitInvoker{prev: client}))

	return client.Run(ctx, func(ctx context.Context) error {
		if err := u.suite.RetryAuthenticate(ctx, client.Auth()); err != nil {
			return xerrors.Errorf("authenticate: %w", err)
		}

		peer, err := sender.Resolve(u.username).AsInputPeer(ctx)
		if err != nil {
			return xerrors.Errorf("resolve bot username %q: %w", u.username, err)
		}

		for _, line := range u.text {
			time.Sleep(2 * time.Second)

			_, err = sender.To(peer).Text(ctx, line)
			if flood, err := tgerr.FloodWait(ctx, err); err != nil {
				if flood {
					continue
				}
				return err
			}

			select {
			case gotMessage := <-u.message:
				require.Equal(u.suite.TB, line, gotMessage)
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	})
}
