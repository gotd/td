package e2etest

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// EchoBot is a simple echo message bot.
type EchoBot struct {
	suite *Suite

	auth chan<- *tg.User
}

// NewEchoBot creates new echo bot.
func NewEchoBot(suite *Suite, auth chan<- *tg.User) EchoBot {
	return EchoBot{
		suite: suite,
		auth:  auth,
	}
}

// Run setups and starts echo bot.
func (b EchoBot) Run(ctx context.Context) error {
	logger := b.suite.Log.Named("echo user")

	dispatcher := tg.NewUpdateDispatcher()
	client, err := b.suite.Client(logger, dispatcher.Handle)
	if err != nil {
		return xerrors.Errorf("create client: %w", err)
	}

	dispatcher.OnNewMessage(func(ctx tg.UpdateContext, u *tg.UpdateNewMessage) error {
		switch m := u.Message.(type) {
		case *tg.Message:
			switch peer := m.PeerID.(type) {
			case *tg.PeerUser:
				user := ctx.Users[peer.UserID]
				logger.With(
					zap.String("text", m.Message),
					zap.Int("user_id", user.ID),
					zap.String("user_first_name", user.FirstName),
					zap.String("username", user.Username),
				).Info("Got message")

				randomID, err := client.RandInt64()
				if err != nil {
					return err
				}
				p := &tg.InputPeerUser{
					UserID:     user.ID,
					AccessHash: user.AccessHash,
				}
				return client.SendMessage(ctx, &tg.MessagesSendMessageRequest{
					RandomID: randomID,
					Message:  m.Message,
					Peer:     p,
				})
			}
		}

		return nil
	})

	auth, err := client.AuthStatus(ctx)
	if err != nil {
		return xerrors.Errorf("get auth status: %w", err)
	}
	logger.With(zap.Bool("authorized", auth.Authorized)).Info("Auth status")

	if err := b.suite.Authenticate(ctx, client); err != nil {
		return xerrors.Errorf("authenticate: %w", err)
	}

	me, err := client.Self(ctx)
	if err != nil {
		return xerrors.Errorf("get self: %w", err)
	}
	logger.With(
		zap.String("user", me.Username),
		zap.Int("id", me.ID),
	).Info("logged")
	b.auth <- me

	<-ctx.Done()
	logger.Info("Shutting down")
	if err := client.Close(); err != nil {
		return err
	}
	logger.Info("Graceful shutdown completed")
	return nil
}
