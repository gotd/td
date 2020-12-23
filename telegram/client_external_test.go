package telegram_test

import (
	"context"
	"crypto/rand"
	"github.com/gotd/td/transport"
	"testing"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/tgflow"
)

func testTransport(transport telegram.Transport) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
			Addr:      telegram.AddrTest,
			Transport: transport,
		})

		if err := client.Connect(ctx); err != nil {
			t.Fatal(err)
		}

		defer func() {
			_ = client.Close()
		}()

		if err := tgflow.NewAuth(tgflow.TestAuth(rand.Reader, 2), telegram.SendCodeOptions{}).Run(ctx, client); err != nil {
			t.Fatal(err)
		}

		if _, err := client.Self(ctx); err != nil {
			t.Fatal(err)
		}
	}
}

func TestExternalE2EConnect(t *testing.T) {
	t.Run("abridged", testTransport(transport.Abridged(nil)))
	t.Run("intermediate", testTransport(transport.Intermediate(nil)))
	t.Run("padded intermediate", testTransport(transport.PaddedIntermediate(nil)))
	t.Run("full", testTransport(transport.Full(nil)))
}
