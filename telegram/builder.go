package telegram

import (
	"context"
	"crypto/rand"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/net/proxy"
	"golang.org/x/xerrors"

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
		dir = "./"
	}

	return filepath.Abs(filepath.Join(dir, ".td"))
}

// OptionsFromEnvironment fills unfilled field in opts parameter
// using environment variables.
func OptionsFromEnvironment(opts Options) (Options, error) {
	// Setting up session storage if not provided.
	if opts.SessionStorage == nil {
		sessionFile, ok := os.LookupEnv("SESSION_FILE")
		if !ok {
			dir, err := sessionDir()
			if err != nil {
				return Options{}, xerrors.Errorf("SESSION_DIR not set or invalid: %w", err)
			}
			sessionFile = filepath.Join(dir, "session.join")
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

// ClientFromEnvironment creates client using environment variables
// but not connects to server.
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

// BotFromEnvironment creates bot client using environment variables
// connects to server and authenticates it.
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

// TestClient creates and authenticates user telegram.Client
// using Telegram staging server.
func TestClient(ctx context.Context, opts Options, cb func(ctx context.Context, client *Client) error) error {
	if opts.Addr == "" {
		opts.Addr = AddrTest
	}

	client := NewClient(TestAppID, TestAppHash, opts)
	return client.Run(ctx, func(ctx context.Context) error {
		if err := NewAuth(
			TestAuth(rand.Reader, 2),
			SendCodeOptions{},
		).Run(ctx, client); err != nil {
			return xerrors.Errorf("auth flow: %w", err)
		}

		return cb(ctx, client)
	})
}
