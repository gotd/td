package e2etest

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// EchoBot is a simple echo message bot.
type EchoBot struct {
	Suite
	token string

	stop chan struct{}
}

// NewEchoBot creates new echo bot.
func NewEchoBot(suite Suite, token string) EchoBot {
	return EchoBot{
		Suite: suite,
		token: token,
		stop:  make(chan struct{}),
	}
}

// Run setups and starts echo bot.
func (b EchoBot) Run(ctx context.Context) error {
	logger := createLogger("bot")
	defer func() { _ = logger.Sync() }()

	dispatcher := tg.NewUpdateDispatcher()
	client := b.Suite.Client(logger, dispatcher.Handle)
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

	err := client.Connect(ctx)
	if err != nil {
		return xerrors.Errorf("failed to connect: %w", err)
	}
	logger.Info("Client started.")

	auth, err := client.AuthStatus(ctx)
	if err != nil {
		return xerrors.Errorf("failed to get auth status: %w", err)
	}
	logger.With(zap.Bool("authorized", auth.Authorized)).Info("Auth status")
	if !auth.Authorized {
		if err := client.AuthBot(ctx, b.token); err != nil {
			return xerrors.Errorf("failed to perform bot login: %w", err)
		}
		logger.Info("Bot login ok")
	}

	// Using tg.Client for directly calling RPC.
	raw := tg.NewClient(client)

	// Getting state is required to process updates in your code.
	// Currently missed updates are not processed, so only new
	// messages will be handled.
	state, err := raw.UpdatesGetState(ctx)
	if err != nil {
		return xerrors.Errorf("failed to get state: %w", err)
	}
	logger.Sugar().Infof("Got state: %+v", state)

	<-b.stop
	logger.Info("Shutting down")
	if err := client.Close(ctx); err != nil {
		return err
	}
	logger.Info("Graceful shutdown completed")
	return nil
}

// Stop stops bot.
func (b EchoBot) Stop() {
	close(b.stop)
}
