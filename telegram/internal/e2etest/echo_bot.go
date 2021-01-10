package e2etest

import (
	"context"
	"strconv"

	"github.com/cenkalti/backoff/v4"
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
	logger := b.suite.Log.Named("echobot")

	dispatcher := tg.NewUpdateDispatcher()
	client := b.suite.Client(logger, dispatcher)
	dispatcher.OnNewMessage(func(ctx tg.UpdateContext, u *tg.UpdateNewMessage) error {
		if m, ok := u.Message.(interface{ GetMessage() string }); ok {
			logger.Named("dispatcher").With(zap.String("message", m.GetMessage())).Info("Got new message update")
		}

		switch m := u.Message.(type) {
		case *tg.Message:
			switch peer := m.PeerID.(type) {
			case *tg.PeerUser:
				user := ctx.Users[peer.UserID]
				logger.Info("Got message", zap.String("text", m.Message),
					zap.Int("user_id", user.ID),
					zap.String("user_first_name", user.FirstName),
					zap.String("username", user.Username))

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

	return client.Run(ctx, func(ctx context.Context) error {
		defer close(b.auth)

		auth, err := client.AuthStatus(ctx)
		if err != nil {
			return xerrors.Errorf("get auth status: %w", err)
		}
		logger.Info("Auth status", zap.Bool("authorized", auth.Authorized))

		if err := b.suite.RetryAuthenticate(ctx, backoff.NewExponentialBackOff(), client); err != nil {
			return xerrors.Errorf("authenticate: %w", err)
		}

		me, err := client.Self(ctx)
		if err != nil {
			return xerrors.Errorf("get self: %w", err)
		}

		raw := tg.NewClient(client)
		_, err = raw.AccountUpdateUsername(ctx, "echobot"+strconv.Itoa(me.ID))
		if err != nil {
			return xerrors.Errorf("update username: %w", err)
		}

		me, err = client.Self(ctx)
		if err != nil {
			return xerrors.Errorf("get self: %w", err)
		}
		logger.Info("Logged in", zap.String("user", me.Username),
			zap.Int("id", me.ID))

		select {
		case b.auth <- me:
		case <-ctx.Done():
			return ctx.Err()
		}

		<-ctx.Done()
		return nil
	})
}
