package peers_test

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
)

func ExampleChannel_RecommendedChannels() {
	logger := zap.NewExample()

	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger: logger.Named("client"),
	})
	if err != nil {
		panic(err)
	}
	peerManager := peers.Options{Logger: logger}.Build(client.API())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := client.Run(ctx, func(ctx context.Context) error {
		if err := peerManager.Init(ctx); err != nil {
			return err
		}

		p, err := peerManager.Resolve(ctx, "telegram")
		if err != nil {
			return err
		}
		ch, ok := p.(peers.Channel)
		if !ok {
			return fmt.Errorf("%q is not a channel", "telegram")
		}

		rec, err := ch.RecommendedChannels(ctx)
		if err != nil {
			return err
		}

		// rec.Count is the total number of recommendations available.
		// For non-Premium accounts the server returns only a subset, so
		// len(rec.Channels) may be smaller than rec.Count.
		fmt.Printf("got %d of %d recommended channels\n", len(rec.Channels), rec.Count)
		for _, c := range rec.Channels {
			fmt.Println(c.VisibleName())
		}
		return nil
	}); err != nil {
		panic(err)
	}
}
