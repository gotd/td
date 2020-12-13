package telegram

import (
	"context"
	"crypto/rsa"
	"strings"
	"testing"
	"time"

	"github.com/gotd/td/telegram/internal/tgtest"
)

func TestClient(t *testing.T) {
	srv := tgtest.NewServer(nil)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	client := NewClient(1, "hash", Options{
		PublicKeys: []*rsa.PublicKey{srv.Key()},
		Addr:       srv.Listener.Addr().String(),
	})

	err := client.Connect(ctx)
	if err == nil {
		t.Error("expected non-nil error")
	} else if !strings.Contains(err.Error(), "nonce mismatch") {
		t.Errorf("expected nonce mismatch, got: %v", err)
	}
}
