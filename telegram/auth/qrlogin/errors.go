package qrlogin

import (
	"errors"
	"fmt"

	"github.com/gotd/td/tg"
)

// ErrAlreadyAuthenticated indicates that user is already authenticated
// and no new QR token is needed.
var ErrAlreadyAuthenticated = errors.New("already authenticated")

// MigrationNeededError reports that Telegram requested DC migration to continue login.
type MigrationNeededError struct {
	MigrateTo *tg.AuthLoginTokenMigrateTo

	// Tried indicates that the migration was attempted.
	//
	// Deprecated: do not use. QR login uses migrate function passed via
	// options.
	Tried bool
}

// Error implements error.
func (m *MigrationNeededError) Error() string {
	return fmt.Sprintf("migration to %d needed", m.MigrateTo.DCID)
}
