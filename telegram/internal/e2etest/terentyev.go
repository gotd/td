package e2etest

import (
	"context"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

// User is a simple user bot.
type User struct {
	suite    *Suite
	text     []string
	username string

	message chan string
}

// NewUser creates new User bot.
func NewUser(suite *Suite, text []string, username string) User {
	return User{
		suite:    suite,
		text:     text,
		username: username,

		message: make(chan string),
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

// Run setups and starts user bot.
func (u User) Run(ctx context.Context) error {
	logger := u.suite.Log.Named("terentyev")
	defer func() { _ = logger.Sync() }()

	dispatcher := tg.NewUpdateDispatcher()
	client := u.suite.Client(logger, dispatcher)
	dispatcher.OnNewMessage(func(ctx tg.UpdateContext, update *tg.UpdateNewMessage) error {
		if m, ok := update.Message.(interface{ GetMessage() string }); ok {
			logger.Named("dispatcher").With(zap.String("message", m.GetMessage())).
				Info("Got new message update")
		}

		expectedMsgText := <-u.message
		msg, ok := update.Message.(*tg.Message)
		if !ok {
			return xerrors.Errorf("unexpected type %T", update.Message)
		}

		require.Equal(u.suite.TB, expectedMsgText, msg.Message)
		return nil
	})

	return client.Run(ctx, func(ctx context.Context) error {
		logger.Info("Client started")

		auth, err := client.AuthStatus(ctx)
		if err != nil {
			return xerrors.Errorf("get auth status: %w", err)
		}
		logger.Info("Auth status", zap.Bool("authorized", auth.Authorized))
		if err := u.suite.RetryAuthenticate(ctx, backoff.NewExponentialBackOff(), client); err != nil {
			return xerrors.Errorf("authenticate: %w", err)
		}

		peer, err := u.resolveBotPeer(ctx, client)
		if err != nil {
			return xerrors.Errorf("resolve bot username %q: %w", u.message, err)
		}

		for _, message := range u.text {
			randomID, err := client.RandInt64()
			if err != nil {
				return err
			}

			err = client.SendMessage(ctx, &tg.MessagesSendMessageRequest{
				RandomID: randomID,
				Message:  message,
				Peer: &tg.InputPeerUser{
					UserID:     peer.ID,
					AccessHash: peer.AccessHash,
				},
			})
			if err != nil {
				return err
			}

			select {
			case u.message <- message:
			case <-ctx.Done():
				break
			}
		}

		logger.Info("Shutting down")
		return nil
	})
}
