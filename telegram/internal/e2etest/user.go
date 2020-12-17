package e2etest

import (
	"context"
	"crypto/rand"

	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/tgflow"
	"github.com/gotd/td/tg"
)

// User is a simple user bot.
type User struct {
	Suite
	text  []string
	botID *tg.User

	message chan string
	stop    chan struct{}
}

// NewUser creates new User bot.
func NewUser(suite Suite, text []string, botID *tg.User) User {
	return User{
		Suite: suite,
		text:  text,
		botID: botID,

		message: make(chan string, 1),
		stop:    make(chan struct{}),
	}
}

// Run setups and starts user bot.
func (u User) Run(ctx context.Context) error {
	logger := createLogger("user")
	defer func() { _ = logger.Sync() }()

	dispatcher := tg.NewUpdateDispatcher()
	client := u.Suite.Client(logger, dispatcher.Handle)
	dispatcher.OnNewMessage(func(ctx tg.UpdateContext, update *tg.UpdateNewMessage) error {
		expectedMsgText := <-u.message
		msg, ok := update.Message.(*tg.Message)
		if !ok {
			return xerrors.Errorf("unexpected type %T", update.Message)
		}

		if msg.Message != expectedMsgText {
			u.TB.Errorf("expected %v, got %v", expectedMsgText, msg.Message)
		}
		return nil
	})

	err := client.Connect(ctx)
	if err != nil {
		return xerrors.Errorf("failed to connect: %w", err)
	}
	defer func() {
		_ = client.Close(ctx)
	}()
	logger.Info("Client started.")

	err = tgflow.NewAuth(
		tgflow.TestAuth(rand.Reader, u.Suite.dcID),
		telegram.SendCodeOptions{},
	).Run(ctx, client)
	if err != nil {
		return xerrors.Errorf("failed to authenticate: %w", err)
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

	return nil
}

// Stop stops bot.
func (u User) Stop() {
	close(u.stop)
}
