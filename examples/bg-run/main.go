// Binary bg-run implements alternative to Run pattern.
package main

import (
	"context"

	"go.uber.org/zap"

	"github.com/gotd/contrib/bg"

	"github.com/nnqq/td/examples"
	"github.com/nnqq/td/telegram"
)

func main() {
	// Some users find explicit client.Run(ctx, f) pattern not very convenient.
	//
	// However, it is possible to implement wrapper and use classic "Connect"
	// pattern instead.
	//
	// The `contrib/bg` package is example implementation of such pattern.
	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		client, err := telegram.ClientFromEnvironment(telegram.Options{
			Logger: log,
		})
		if err != nil {
			return err
		}

		// bg.Connect will call Run in background.
		// Call stop() to disconnect and release resources.
		stop, err := bg.Connect(client)
		if err != nil {
			return err
		}
		defer func() { _ = stop() }()

		// Now you can use client.
		if _, err := client.Auth().Status(ctx); err != nil {
			return err
		}

		return nil
	})
}
