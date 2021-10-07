package telegram_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/tdsync"
	"github.com/nnqq/td/session"
	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/dcs"
)

type mtg struct {
	path string
	addr string
}

type signalWriter struct {
	io.Writer
	wait *tdsync.Ready
}

func (s signalWriter) Write(p []byte) (n int, err error) {
	s.wait.Signal()
	return s.Writer.Write(p)
}

func (m mtg) run(ctx context.Context, secret string, out, err io.Writer, wait *tdsync.Ready) error {
	cmd := exec.CommandContext(ctx, m.path, "run", "--bind", m.addr, "-4", m.addr, secret)
	cmd.Stdout = signalWriter{Writer: out, wait: wait}
	cmd.Stderr = signalWriter{Writer: err, wait: wait}
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

func testMTProxy(secretType string, m mtg, storage session.Storage) func(t *testing.T) {
	return func(t *testing.T) {
		a := require.New(t)
		logger := zaptest.NewLogger(t)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		secret, err := m.generateSecret(ctx, secretType)
		a.NoError(err)

		// Store mtg logs to buffer and print it only if test failed.
		w := &bytes.Buffer{}
		t.Cleanup(func() {
			if t.Failed() {
				_, _ = io.Copy(os.Stdout, w)
			}
		})

		g := tdsync.NewCancellableGroup(ctx)
		ready := tdsync.NewReady()
		g.Go(func(ctx context.Context) error {
			err := m.run(ctx, hex.EncodeToString(secret), w, w, ready)
			select {
			case <-ctx.Done():
				return nil
			default:
				return err
			}
		})
		g.Go(func(ctx context.Context) error {
			defer g.Cancel()
			select {
			case <-ready.Ready():
			case <-ctx.Done():
				return ctx.Err()
			}

			resolver, err := dcs.MTProxy(m.addr, secret, dcs.MTProxyOptions{})
			if err != nil {
				return err
			}

			return tryConnect(ctx, telegram.Options{
				Resolver:       resolver,
				Logger:         logger,
				SessionStorage: storage,
				DCList:         dcs.Test(),
			})
		})

		a.NoError(g.Wait())
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

	// To re-use session.
	storage := &session.StorageMemory{}
	m := mtg{path: mtgPath, addr: addr}
	for _, secretType := range []string{"simple", "secured", "tls"} {
		t.Run(strings.Title(secretType), testMTProxy(secretType, m, storage))
	}
}
