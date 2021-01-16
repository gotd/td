package telegram

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	"golang.org/x/net/proxy"
	"golang.org/x/xerrors"

	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/session"
	"github.com/gotd/td/transport"
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
// SESSION_FILE: path to session file
// SESSION_DIR: path to session directory, if SESSION_FILE is not set
// ALL_PROXY, NO_PROXY: see https://pkg.go.dev/golang.org/x/net/proxy#FromEnvironment
func OptionsFromEnvironment(opts Options) (Options, error) {
	// Setting up session storage if not provided.
	if opts.SessionStorage == nil {
		sessionFile, ok := os.LookupEnv("SESSION_FILE")
		if !ok {
			dir, err := sessionDir()
			if err != nil {
				return Options{}, xerrors.Errorf("SESSION_DIR not set or invalid: %w", err)
			}
			sessionFile = filepath.Join(dir, "session.json")
		}

		dir, _ := filepath.Split(sessionFile)
		if err := os.MkdirAll(dir, 0600); err != nil {
			return Options{}, xerrors.Errorf("session dir creation: %w", err)
		}

		opts.SessionStorage = &session.FileStorage{
			Path: sessionFile,
		}
	}

	if opts.Transport == nil {
		opts.Transport = transport.Intermediate(transport.DialFunc(proxy.Dial))
	}

	return opts, nil
}

// ClientFromEnvironment creates client using OptionsFromEnvironment
// but does not connect to server.
//
// Variables:
// APP_ID — app_id of Telegram app.
// APP_HASH — app_hash of Telegram app.
func ClientFromEnvironment(opts Options) (*Client, error) {
	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		return nil, xerrors.Errorf("APP_ID not set or invalid: %w", err)
	}

	appHash := os.Getenv("APP_HASH")
	if appHash == "" {
		return nil, xerrors.New("no APP_HASH provided")
	}

	opts, err = OptionsFromEnvironment(opts)
	if err != nil {
		return nil, err
	}

	return NewClient(appID, appHash, opts), nil
}

// BotFromEnvironment creates bot client using ClientFromEnvironment
// connects to server and authenticates it.
//
// Variables:
// BOT_TOKEN — token from BotFather.
func BotFromEnvironment(ctx context.Context, opts Options, cb func(ctx context.Context, client *Client) error) error {
	client, err := ClientFromEnvironment(opts)
	if err != nil {
		return err
	}

	return client.Run(ctx, func(ctx context.Context) error {
		status, err := client.AuthStatus(ctx)
		if err != nil {
			return xerrors.Errorf("auth status: %w", err)
		}

		if !status.Authorized {
			if err := client.AuthBot(ctx, os.Getenv("BOT_TOKEN")); err != nil {
				return xerrors.Errorf("login: %w", err)
			}
		}

		return cb(ctx, client)
	})
}

func retry(ctx context.Context, cb func(ctx context.Context) error) error {
	b := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)

	return backoff.Retry(func() error {
		if err := cb(ctx); err != nil {
			var rpcErr *mtproto.Error
			if errors.As(err, &rpcErr) {
				switch rpcErr.Type {
				case "NEED_MEMBER_INVALID",
					"AUTH_KEY_UNREGISTERED",
					"API_ID_PUBLISHED_FLOOD":
					return err
				case "FLOOD_WAIT":
					time.Sleep(time.Duration(rpcErr.Argument) * time.Second)
					return err
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
// using Telegram staging server.
func TestClient(ctx context.Context, opts Options, cb func(ctx context.Context, client *Client) error) error {
	if opts.Addr == "" {
		opts.Addr = AddrTest
	}

	// Sometimes testing server can return "AUTH_KEY_UNREGISTERED" error.
	// It is expected and client implementation is unlikely to cause
	// such errors, so just doing retries using backoff.
	return retry(ctx, func(retryCtx context.Context) error {
		client := NewClient(TestAppID, TestAppHash, opts)
		return client.Run(retryCtx, func(runCtx context.Context) error {
			if err := NewAuth(
				TestAuth(rand.Reader, 2),
				SendCodeOptions{},
			).Run(runCtx, client); err != nil {
				return xerrors.Errorf("auth flow: %w", err)
			}

			return cb(runCtx, client)
		})
	})
}
