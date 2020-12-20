package telegram_test

import (
	"context"
	"crypto/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"golang.org/x/net/proxy"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/tgflow"
	"github.com/gotd/td/transport"
)

func TestExternalE2EConnect(t *testing.T) {
	if ok, _ := strconv.ParseBool(os.Getenv("GOTD_TEST_EXTERNAL")); !ok {
		t.Skip("Skipped. Set GOTD_TEST_EXTERNAL=1 to enable external e2e test.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
		Addr:      telegram.AddrTest,
		Transport: transport.Intermediate(transport.DialFunc(proxy.Dial)),
	})

	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = client.Close(ctx)
	}()

	if err := tgflow.NewAuth(tgflow.TestAuth(rand.Reader, 2), telegram.SendCodeOptions{}).Run(ctx, client); err != nil {
		t.Fatal(err)
	}
}
