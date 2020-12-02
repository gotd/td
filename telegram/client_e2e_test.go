package telegram

import (
	"context"
	"crypto/rsa"
	"strings"
	"testing"
	"time"

	"github.com/gotd/td/telegram/internal/tgtest"
)

func TestDial(t *testing.T) {
	srv := tgtest.NewServer(nil)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	client, err := Dial(ctx, Options{
		AppID:   1,
		AppHash: "hash",

		PublicKeys: []*rsa.PublicKey{srv.Key()},
		Addr:       srv.Listener.Addr().String(),
	})
	if client != nil {
		t.Error("expected nil client")
	}
	if err == nil {
		t.Error("expected non-nil error")
	} else if !strings.Contains(err.Error(), "nonce mismatch") {
		t.Error("expected nonce mismatch")
	}
}
