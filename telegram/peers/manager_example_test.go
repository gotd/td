package peers_test

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
)

func ExampleManager() {
	logger := zap.NewExample()

	var (
		dispatcher = tg.NewUpdateDispatcher()
		h          telegram.UpdateHandler
	)
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger: logger.Named("client"),
		UpdateHandler: telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
			return h.Handle(ctx, u)
		}),
	})
	if err != nil {
		panic(err)
	}
	peerManager := peers.Options{
		Logger: logger,
	}.Build(client.API())
	gaps := updates.New(updates.Config{
		Handler:      dispatcher,
		AccessHasher: peerManager,
		Logger:       logger.Named("gaps"),
	})
	h = peerManager.UpdateHook(gaps)

	if err := client.Run(context.TODO(), func(ctx context.Context) error {
		if err := peerManager.Init(ctx); err != nil {
			return err
		}
		u, err := peerManager.Self(ctx)
		if err != nil {
			return err
		}

		_, isBot := u.ToBot()
		if err := gaps.Auth(ctx, client.API(), u.ID(), isBot, false); err != nil {
			return err
		}
		defer gaps.Logout()

		p, err := peerManager.Resolve(ctx, "durov")
		if err != nil {
			return err
		}

		username, _ := p.Username()
		fmt.Println(username)
		return nil
	}); err != nil {
		panic(err)
	}
}
