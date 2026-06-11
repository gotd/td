package srpguard_test

import (
	"context"

	"github.com/awnumar/memguard"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/auth/srpguard"
)

// This example shows how to supply a 2FA password from protected memory instead
// of a Go string, so the plaintext is locked, never swapped to disk, and wiped
// after the SRP answer is computed.
func ExampleLockedBuffer() {
	// secret is read from a prompt/keyring into a byte slice; memguard takes
	// ownership of it and wipes the original.
	secret := []byte("correct horse battery staple")
	buf := memguard.NewBufferFromBytes(secret)

	// client is obtained via telegramClient.Auth().
	var client *auth.Client
	_, _ = client.PasswordWith(context.Background(), srpguard.LockedBuffer(buf))
}
