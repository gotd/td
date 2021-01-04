package e2etest

import (
	"context"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// User is a simple user bot.
type User struct {
	suite *Suite
	text  []string
	botID *tg.User

	message chan string
}

// NewUser creates new User bot.
func NewUser(suite *Suite, text []string, botID *tg.User) User {
	return User{
		suite: suite,
		text:  text,
		botID: botID,

		message: make(chan string),
	}
}

// Run setups and starts user bot.
func (u User) Run(ctx context.Context) error {
	logger := u.suite.Log.Named("terentyev")
	defer func() { _ = logger.Sync() }()

	dispatcher := tg.NewUpdateDispatcher()
	client := u.suite.Client(logger, dispatcher.Handle)
	dispatcher.OnNewMessage(func(ctx tg.UpdateContext, update *tg.UpdateNewMessage) error {
		expectedMsgText := <-u.message
		msg, ok := update.Message.(*tg.Message)
		if !ok {
			return xerrors.Errorf("unexpected type %T", update.Message)
		}

		require.Equal(u.suite.TB, expectedMsgText, msg.Message)
		return nil
	})

	runCtx, runCancel := context.WithCancel(ctx)
	defer runCancel()

	g, gCtx := errgroup.WithContext(runCtx)
	g.Go(func() error {
		return client.Run(gCtx)
	})

	logger.Info("Client started.")

	auth, err := client.AuthStatus(ctx)
	if err != nil {
		return xerrors.Errorf("get auth status: %w", err)
	}
	logger.With(zap.Bool("authorized", auth.Authorized)).Info("Auth status")
	if err := u.suite.Authenticate(ctx, client); err != nil {
		return xerrors.Errorf("authenticate: %w", err)
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
				UserID:     u.botID.ID,
				AccessHash: u.botID.AccessHash,
			},
		})
		if err != nil {
			return err
		}

		u.message <- message
	}

	logger.Info("Shutting down")
	runCancel()
	if err := g.Wait(); err != nil {
		return xerrors.Errorf("wait: %w", err)
	}
	logger.Info("Graceful shutdown completed")
	return nil
}
