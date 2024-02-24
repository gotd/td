package telegram

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"golang.org/x/net/proxy"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tgerr"
)

func sessionDir() (string, error) {
	dir, ok := os.LookupEnv("SESSION_DIR")
	if ok {
		return filepath.Abs(dir)
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		dir = "."
	}

	return filepath.Abs(filepath.Join(dir, ".td"))
}

// OptionsFromEnvironment fills unfilled field in opts parameter
// using environment variables.
//
// Variables:
//
//	SESSION_FILE:        path to session file
//	SESSION_DIR:         path to session directory, if SESSION_FILE is not set
//	ALL_PROXY, NO_PROXY: see https://pkg.go.dev/golang.org/x/net/proxy#FromEnvironment
func OptionsFromEnvironment(opts Options) (Options, error) {
	// Setting up session storage if not provided.
	if opts.SessionStorage == nil {
		sessionFile, ok := os.LookupEnv("SESSION_FILE")
		if !ok {
			dir, err := sessionDir()
			if err != nil {
				return Options{}, errors.Wrap(err, "SESSION_DIR not set or invalid")
			}
			sessionFile = filepath.Join(dir, "session.json")
		}

		dir, _ := filepath.Split(sessionFile)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return Options{}, errors.Wrap(err, "session dir creation")
		}

		opts.SessionStorage = &session.FileStorage{
			Path: sessionFile,
		}
	}

	if opts.Resolver == nil {
		opts.Resolver = dcs.Plain(dcs.PlainOptions{
			Dial: proxy.Dial,
		})
	}

	return opts, nil
}

// ClientFromEnvironment creates client using OptionsFromEnvironment
// but does not connect to server.
//
// Variables:
//
//	APP_ID:   app_id of Telegram app.
//	APP_HASH: app_hash of Telegram app.
func ClientFromEnvironment(opts Options) (*Client, error) {
	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		return nil, errors.Wrap(err, "APP_ID not set or invalid")
	}

	appHash := os.Getenv("APP_HASH")
	if appHash == "" {
		return nil, errors.New("no APP_HASH provided")
	}

	opts, err = OptionsFromEnvironment(opts)
	if err != nil {
		return nil, err
	}

	return NewClient(appID, appHash, opts), nil
}

func retry(ctx context.Context, logger *zap.Logger, cb func(ctx context.Context) error) error {
	b := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)

	// List of known retryable RPC error types.
	retryableErrors := []string{
		"NEED_MEMBER_INVALID",
		"AUTH_KEY_UNREGISTERED",
		"API_ID_PUBLISHED_FLOOD",
	}

	return backoff.Retry(func() error {
		if err := cb(ctx); err != nil {
			logger.Warn("TestClient run failed", zap.Error(err))

			if tgerr.Is(err, retryableErrors...) {
				return err
			}
			if timeout, ok := AsFloodWait(err); ok {
				timer := clock.System.Timer(timeout + 1*time.Second)
				defer clock.StopTimer(timer)

				select {
				case <-timer.C():
					return err
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				// Possibly server closed connection.
				return err
			}

			return backoff.Permanent(err)
		}

		return nil
	}, b)
}

// TestClient creates and authenticates user telegram.Client
// using Telegram test server.
func TestClient(ctx context.Context, opts Options, cb func(ctx context.Context, client *Client) error) error {
	if opts.DC == 0 {
		opts.DC = 2
	}
	if opts.DCList.Zero() {
		opts.DCList = dcs.Test()
	}

	logger := zap.NewNop()
	if opts.Logger != nil {
		logger = opts.Logger.Named("test")
	}

	// Sometimes testing server can return "AUTH_KEY_UNREGISTERED" error.
	// It is expected and client implementation is unlikely to cause
	// such errors, so just doing retries using backoff.
	return retry(ctx, logger, func(retryCtx context.Context) error {
		client := NewClient(TestAppID, TestAppHash, opts)
		return client.Run(retryCtx, func(runCtx context.Context) error {
			if err := client.Auth().IfNecessary(runCtx, auth.NewFlow(
				auth.Test(crypto.DefaultRand(), opts.DC),
				auth.SendCodeOptions{},
			)); err != nil {
				return errors.Wrap(err, "auth flow")
			}

			return cb(runCtx, client)
		})
	})
}
