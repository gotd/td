package telegram_test

import (
	"context"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/transport"
)

func TestExternalE2EMTProxy(t *testing.T) {
	addr, ok := os.LookupEnv("GOTD_MTPROXY_ADDR")
	if !ok {
		t.Skip("Skipped. Set GOTD_MTPROXY_ADDR to enable external e2e mtproxy test.")
	}

	secret, err := hex.DecodeString(os.Getenv("GOTD_MTPROXY_SECRET"))
	if err != nil {
		t.Fatal("secret parsing failed", err)
	}

	trp, err := transport.MTProxy(nil, 2, secret)
	if err != nil {
		t.Fatal("secret invalid", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err = telegram.TestClient(ctx, telegram.Options{
		Addr:      addr,
		Transport: trp,
	}, func(ctx context.Context, client *telegram.Client) error {
		if _, err := client.Self(ctx); err != nil {
			return xerrors.Errorf("self: %w", err)
		}

		return nil
	})
	require.NoError(t, err)
}
