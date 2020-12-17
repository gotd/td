package e2etest

import (
	"context"
	"crypto/rand"
	"strings"

	"go.uber.org/zap"

	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/tgflow"
	"github.com/gotd/td/tg"
)

// BotCreator is user bot which creates plain bots.
type BotCreator struct {
	Suite
	client    *telegram.Client
	botFather *tg.User

	tokenMsg chan string
}

// NewBotCreator create new BotCreator.
func NewBotCreator(suite Suite) BotCreator {
	return BotCreator{
		Suite:    suite,
		tokenMsg: make(chan string, 1),
	}
}

// Connect setups and starts user bot.
func (u *BotCreator) Connect(ctx context.Context) error {
	logger := createLogger("box_creator")
	defer func() { _ = logger.Sync() }()

	dispatcher := tg.NewUpdateDispatcher()
	u.client = u.Suite.Client(logger, dispatcher.Handle)
	dispatcher.OnNewMessage(func(ctx tg.UpdateContext, update *tg.UpdateNewMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok {
			return xerrors.Errorf("unexpected type %T", update.Message)
		}
		logger.Named("message").
			With(zap.String("message", msg.Message)).
			Info("Got message")

		if strings.Contains(msg.Message, "Use this token to access the HTTP API:") {
			u.tokenMsg <- msg.Message
		}
		return nil
	})

	err := u.client.Connect(ctx)
	if err != nil {
		return xerrors.Errorf("failed to connect: %w", err)
	}
	logger.Info("Client started.")

	err = tgflow.NewAuth(
		tgflow.TestAuth(rand.Reader, u.Suite.dcID),
		telegram.SendCodeOptions{},
	).Run(ctx, u.client)
	if err != nil {
		return xerrors.Errorf("failed to authenticate: %w", err)
	}

	return nil
}

// ResolveUsername resolves username using Telegram API.
func (u BotCreator) ResolveUsername(ctx context.Context, username string) (*tg.User, error) {
	if u.botFather != nil {
		return u.botFather, nil
	}

	raw := tg.NewClient(u.client)
	p, err := raw.ContactsResolveUsername(ctx, username)
	if err != nil {
		return nil, xerrors.Errorf("failed to resolve username %s: %w", username, err)
	}

	if len(p.Users) < 1 {
		return nil, xerrors.Errorf("result is empty")
	}

	user, ok := p.Users[0].(*tg.User)
	if !ok {
		return nil, xerrors.Errorf("unexpected type %T", p.Users[0])
	}

	return user, nil
}

func (u BotCreator) runCommands(ctx context.Context, commands ...string) error {
	father, err := u.ResolveUsername(ctx, "BotFather")
	if err != nil {
		return err
	}

	for _, command := range commands {
		id, err := u.client.RandInt64()
		if err != nil {
			return xerrors.Errorf("random_id generation failed: %w", err)
		}

		err = u.client.SendMessage(ctx, &tg.MessagesSendMessageRequest{
			Peer: &tg.InputPeerUser{
				UserID:     father.ID,
				AccessHash: father.AccessHash,
			},
			Message:  command,
			RandomID: id,
		})
		if err != nil {
			return xerrors.Errorf("send message failed while creating bot: %w", err)
		}
	}

	return nil
}

func parseTokenMessage(msg string) (string, error) {
	tokenMsg := msg // copy

	const before = "Use this token to access the HTTP API:"
	index := strings.Index(tokenMsg, before)
	if index < 0 {
		return "", xerrors.Errorf("invalid token respond: %s", msg)
	}

	tokenMsg = tokenMsg[index+len(before):]
	index = strings.Index(tokenMsg, "Keep your token")
	if index < 0 {
		return "", xerrors.Errorf("invalid token respond: %s", msg)
	}

	tokenMsg = tokenMsg[:index]
	if len(tokenMsg) < 2 {
		return "", xerrors.Errorf("invalid token respond: %s", msg)
	}
	tokenMsg = tokenMsg[1 : len(tokenMsg)-1] // unquote ``

	return tokenMsg, nil
}

// CreateBot sends commands to the @BotFather to create bot and returns token.
func (u BotCreator) CreateBot(ctx context.Context, name string) (string, error) {
	err := u.runCommands(ctx,
		"/cancel",
		"/newbot",
		name,
		name,
	)
	if err != nil {
		return "", err
	}

	return parseTokenMessage(<-u.tokenMsg)
}

// DeleteBot sends commands to the @BotFather to delete bot created by CreateBot method.
func (u BotCreator) DeleteBot(ctx context.Context, name string) error {
	return u.runCommands(ctx,
		"/cancel",
		"/deletebot",
		"@"+name,
		"Yes, I am totally sure.",
	)
}

// Close stops and closes BotCreator.
func (u BotCreator) Close(ctx context.Context) error {
	if u.client != nil {
		return u.client.Close(ctx)
	}

	return nil
}
