// Binary secure-password logs in with phone + code and supplies the 2FA
// password from protected memory (memguard) instead of a Go string, so the
// plaintext is locked, never swapped to disk, and wiped after the SRP answer is
// computed (gotd/td#755).
//
// Usage:
//
//	APP_ID=... APP_HASH=... PHONE=+1234567890 SESSION_FILE=session.json \
//	    go run ./examples/secure-password
package main

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/awnumar/memguard"
	"go.uber.org/zap"
	"golang.org/x/term"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/auth/srpguard"
	"github.com/gotd/td/tg"
)

func main() {
	// Wipe all protected memory if the process is interrupted.
	memguard.CatchInterrupt()
	defer memguard.Purge()

	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		// Read the 2FA password into protected memory once, up front.
		// NewEnclave copies the input into an encrypted enclave and wipes it.
		fmt.Fprint(os.Stderr, "Enter 2FA password: ")
		secret, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr)
		enclave := memguard.NewEnclave(secret)

		client, err := telegram.ClientFromEnvironment(telegram.Options{Logger: log})
		if err != nil {
			return err
		}

		flow := auth.NewFlow(
			securePassword{
				UserAuthenticator: examples.Terminal{PhoneNumber: os.Getenv("PHONE")},
				enclave:           enclave,
			},
			auth.SendCodeOptions{},
		)

		return client.Run(ctx, func(ctx context.Context) error {
			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				return err
			}
			self, err := client.Self(ctx)
			if err != nil {
				return err
			}
			log.Info("Logged in with secure 2FA password",
				zap.String("user", self.FirstName), zap.Int64("id", self.ID))
			return nil
		})
	})
}

// securePassword extends the terminal authenticator to supply the 2FA password
// from a memguard enclave instead of a string. Because it implements
// auth.PasswordHashProvider, auth.Flow uses PasswordHash and never calls the
// string-based Password fallback.
type securePassword struct {
	auth.UserAuthenticator
	enclave *memguard.Enclave
}

func (s securePassword) PasswordHash(ctx context.Context, p *tg.AccountPassword) (*tg.InputCheckPasswordSRP, error) {
	return srpguard.Enclave(s.enclave)(ctx, p)
}
