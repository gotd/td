// Binary bot-auth-manual implements example of custom session storage and
// manually setting up client options without environment variables.
package main

import (
	"context"
	"flag"
	"sync"

	"go.uber.org/zap"

	"github.com/nnqq/td/examples"
	"github.com/nnqq/td/session"
	"github.com/nnqq/td/telegram"
)

// memorySession implements in-memory session storage.
// Goroutine-safe.
type memorySession struct {
	mux  sync.RWMutex
	data []byte
}

// LoadSession loads session from memory.
func (s *memorySession) LoadSession(context.Context) ([]byte, error) {
	if s == nil {
		return nil, session.ErrNotFound
	}

	s.mux.RLock()
	defer s.mux.RUnlock()

	if len(s.data) == 0 {
		return nil, session.ErrNotFound
	}

	cpy := append([]byte(nil), s.data...)

	return cpy, nil
}

// StoreSession stores session to memory.
func (s *memorySession) StoreSession(ctx context.Context, data []byte) error {
	s.mux.Lock()
	s.data = data
	s.mux.Unlock()
	return nil
}

func main() {
	// Grab those from https://my.telegram.org/apps.
	appID := flag.Int("api-id", 0, "app id")
	appHash := flag.String("api-hash", "hash", "app hash")
	// Get it from bot father.
	token := flag.String("token", "", "bot token")
	flag.Parse()

	// Using custom session storage.
	// You can save session to database, e.g. Redis, MongoDB or postgres.
	// See memorySession for implementation details.
	sessionStorage := &memorySession{}

	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		client := telegram.NewClient(*appID, *appHash, telegram.Options{
			SessionStorage: sessionStorage,
			Logger:         log,
		})

		return client.Run(ctx, func(ctx context.Context) error {
			// Checking auth status.
			status, err := client.Auth().Status(ctx)
			if err != nil {
				return err
			}
			// Can be already authenticated if we have valid session in
			// session storage.
			if !status.Authorized {
				// Otherwise, perform bot authentication.
				if _, err := client.Auth().Bot(ctx, *token); err != nil {
					return err
				}
			}

			// All good, manually authenticated.
			log.Info("Done")

			return nil
		})
	})
}
