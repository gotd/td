package telegram_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/dcs"
	"github.com/nnqq/td/tg"
)

func connectViaMTProxy(ctx context.Context) error {
	secret, err := hex.DecodeString(os.Getenv("SECRET"))
	if err != nil {
		return xerrors.Errorf("parse secret: %w", err)
	}

	resolver, err := dcs.MTProxy(
		os.Getenv("PROXY_ADDR"),
		secret,
		dcs.MTProxyOptions{},
	)
	if err != nil {
		return xerrors.Errorf("create MTProxy resolver: %w", err)
	}

	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Resolver: resolver,
	})
	if err != nil {
		return xerrors.Errorf("create client: %w", err)
	}

	return client.Run(ctx, func(ctx context.Context) error {
		cfg, err := tg.NewClient(client).HelpGetConfig(ctx)
		if err != nil {
			return xerrors.Errorf("get config: %w", err)
		}

		fmt.Println("This DC: ", cfg.ThisDC)
		return nil
	})
}

func ExampleClient_mtproxy() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := connectViaMTProxy(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
