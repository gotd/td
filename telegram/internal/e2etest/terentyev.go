package e2etest

import (
	"context"
	"errors"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
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
		logger:   suite.Log.Named("terentyev"),
		message:  make(chan string, 1),
	}
}

func (u User) resolveBotPeer(ctx context.Context, client *telegram.Client) (*tg.User, error) {
	raw := tg.NewClient(client)
	peer, err := raw.ContactsResolveUsername(ctx, u.username)
	if err != nil {
		return nil, err
	}

	users := peer.GetUsers()
	if len(users) != 1 {
		return nil, xerrors.Errorf("expected users field length is equal to 1, got %d", len(users))
	}

	user, ok := users[0].(*tg.User)
	if !ok {
		return nil, xerrors.Errorf("unexpected peer type %T", peer.GetPeer())
	}

	return user, nil
}

func (u User) messageHandler(ctx tg.UpdateContext, update *tg.UpdateNewMessage) error {
	if filterMessage(update) {
		return nil
	}

	if m, ok := update.Message.(interface{ GetMessage() string }); ok {
		u.logger.Named("dispatcher").
			With(zap.String("message", m.GetMessage())).
			Info("Got new message update")
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

	return client.Run(ctx, func(ctx context.Context) error {
		if err := u.suite.RetryAuthenticate(ctx, client); err != nil {
			return xerrors.Errorf("authenticate: %w", err)
		}

		peer, err := u.resolveBotPeer(ctx, client)
		if err != nil {
			return xerrors.Errorf("resolve bot username %q: %w", u.username, err)
		}

		for _, message := range u.text {
			randomID, err := client.RandInt64()
			if err != nil {
				return err
			}

			time.Sleep(2 * time.Second)
			err = client.SendMessage(ctx, &tg.MessagesSendMessageRequest{
				RandomID: randomID,
				Message:  message,
				Peer: &tg.InputPeerUser{
					UserID:     peer.ID,
					AccessHash: peer.AccessHash,
				},
			})
			if err != nil {
				var rpcErr *mtproto.Error
				if !errors.As(err, &rpcErr) || rpcErr.Message != "FLOOD_WAIT" {
					return err
				}
				time.Sleep(time.Duration(rpcErr.Argument) * time.Second)

				continue //
			}

			select {
			case gotMessage := <-u.message:
				require.Equal(u.suite.TB, message, gotMessage)
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	})
}
