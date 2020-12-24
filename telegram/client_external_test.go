package telegram_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gotd/td/transport"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/tgflow"
)

func testTransport(trp telegram.Transport) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
			Addr:      telegram.AddrTest,
			Transport: trp,
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
	if ok, _ := strconv.ParseBool(os.Getenv("GOTD_TEST_EXTERNAL")); !ok {
		t.Skip("Skipped. Set GOTD_TEST_EXTERNAL=1 to enable external e2e test.")
	}

	t.Run("abridged", testTransport(transport.Abridged(nil)))
	t.Run("intermediate", testTransport(transport.Intermediate(nil)))
	t.Run("padded intermediate", testTransport(transport.PaddedIntermediate(nil)))
	t.Run("full", testTransport(transport.Full(nil)))
}

func TestMTProxy(t *testing.T) {
	secret, err := hex.DecodeString("8a96ef6e42a18c21837580cd1c91c5a8")
	if err != nil {
		t.Fatal(err)
	}

	trp := transport.MTProxy(nil, 2, secret)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
		Addr:      "localhost:3128",
		Transport: trp,
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
