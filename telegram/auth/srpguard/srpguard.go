// Package srpguard provides memguard-backed 2FA password handling for
// telegram/auth, keeping the plaintext password in locked, swap-protected
// memory that is wiped after the SRP answer is computed.
//
// It addresses gotd/td#755: a Go string cannot be reliably zeroed, so the 2FA
// password may linger in memory longer than necessary. The helpers here return
// an [auth.PasswordHashFunc] that reads the password from a memguard buffer and
// destroys it before returning.
//
// Usage with the high-level method:
//
//	buf := memguard.NewBufferFromBytes(secret) // takes ownership, wipes secret
//	_, err := client.Auth().PasswordWith(ctx, srpguard.LockedBuffer(buf))
//
// or with an encrypted [memguard.Enclave]:
//
//	_, err := client.Auth().PasswordWith(ctx, srpguard.Enclave(enclave))
//
// This package isolates the memguard dependency from the core auth package.
package srpguard

import (
	"context"

	"github.com/awnumar/memguard"
	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

// errDestroyed is returned when the provided buffer is no longer usable.
var errDestroyed = errors.New("srpguard: password buffer already destroyed")

// LockedBuffer returns an [auth.PasswordHashFunc] that computes the SRP answer
// from a password kept in buf, destroying buf afterwards.
//
// buf is consumed: it is destroyed once the returned function is called (or, if
// it is never called, the caller remains responsible for destroying it).
func LockedBuffer(buf *memguard.LockedBuffer) auth.PasswordHashFunc {
	return func(ctx context.Context, p *tg.AccountPassword) (*tg.InputCheckPasswordSRP, error) {
		defer buf.Destroy()
		if !buf.IsAlive() {
			return nil, errDestroyed
		}
		return auth.PasswordHash(buf.Bytes(), p.SRPID, p.SRPB, p.SecureRandom, p.CurrentAlgo)
	}
}

// Enclave returns an [auth.PasswordHashFunc] that opens enc into a locked
// buffer, computes the SRP answer and destroys the buffer. enc itself remains
// valid and may be reused.
func Enclave(enc *memguard.Enclave) auth.PasswordHashFunc {
	return func(ctx context.Context, p *tg.AccountPassword) (*tg.InputCheckPasswordSRP, error) {
		buf, err := enc.Open()
		if err != nil {
			return nil, errors.Wrap(err, "open enclave")
		}
		defer buf.Destroy()
		return auth.PasswordHash(buf.Bytes(), p.SRPID, p.SRPB, p.SecureRandom, p.CurrentAlgo)
	}
}
