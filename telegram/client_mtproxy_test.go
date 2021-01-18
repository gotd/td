package telegram_test

import (
	"context"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/transport"
)

type mtg struct {
	path string
}

type signalWriter struct {
	io.Writer
	wait *tdsync.Ready
}

func (s signalWriter) Write(p []byte) (n int, err error) {
	s.wait.Signal()
	return s.Writer.Write(p)
}

func (m mtg) run(ctx context.Context, secret, addr string, wait *tdsync.Ready) error {
	cmd := exec.CommandContext(ctx, m.path, "run", "--bind", addr, secret)
	cmd.Stdout = signalWriter{Writer: os.Stdout, wait: wait}
	cmd.Stderr = signalWriter{Writer: os.Stderr, wait: wait}
	cmd.Env = append([]string{"MTG_DEBUG=true", "MTG_TEST_DC=true"}, os.Environ()...)
	return cmd.Run()
}

func (m mtg) generateSecret(ctx context.Context, t string) ([]byte, error) {
	args := []string{"generate-secret"}
	if t == "tls" {
		args = append(args, "-c", "google.com")
	}
	args = append(args, t)

	o, err := exec.CommandContext(ctx, m.path, args...).Output()
	if err != nil {
		return nil, xerrors.Errorf("execute command: %w", err)
	}
	output := strings.TrimSpace(string(o))

	r, err := hex.DecodeString(output)
	if err != nil {
		return nil, xerrors.Errorf("decode secret %q: %w", output, err)
	}

	return r, nil
}

func testMTProxy(secretType, addr string, m mtg) func(t *testing.T) {
	return func(t *testing.T) {
		a := require.New(t)

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		secret, err := m.generateSecret(ctx, secretType)
		a.NoError(err)

		grp := tdsync.NewCancellableGroup(ctx)
		ready := tdsync.NewReady()
		grp.Go(func(groupCtx context.Context) error {
			_ = m.run(groupCtx, hex.EncodeToString(secret), addr, ready)
			return nil
		})
		grp.Go(func(groupCtx context.Context) error {
			defer grp.Cancel()
			<-ready.Ready()

			trp, err := transport.MTProxy(nil, 2, secret)
			if err != nil {
				return err
			}

			return telegram.TestClient(ctx, telegram.Options{
				Addr:      addr,
				Transport: trp,
			}, func(ctx context.Context, client *telegram.Client) error {
				if _, err := client.Self(ctx); err != nil {
					return xerrors.Errorf("self: %w", err)
				}

				return nil
			})
		})

		a.NoError(grp.Wait())
	}
}

func TestExternalE2EMTProxy(t *testing.T) {
	addr, ok := os.LookupEnv("GOTD_MTPROXY_ADDR")
	if !ok {
		t.Skip("Skipped. Set GOTD_MTPROXY_ADDR to enable external e2e mtproxy test.")
	}

	mtgPath, err := exec.LookPath("mtg")
	if err != nil {
		t.Fatal("mtg binary not found", err)
	}

	m := mtg{path: mtgPath}
	for _, secretType := range []string{"simple", "secured", "tls"} {
		t.Run(strings.Title(secretType), testMTProxy(secretType, addr, m))
	}
}
