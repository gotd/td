package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync/atomic"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/crypto"
)

// SessionStorage is secure persistent storage for client session.
//
// NB: Implementation security is important, attacker can use not only for
// connecting as authenticated user or bot, but even decrypting previous
// messages in some situations.
type SessionStorage interface {
	LoadSession(ctx context.Context) ([]byte, error)
	StoreSession(ctx context.Context, data []byte) error
}

// ErrSessionNotFound means that session is not found in storage.
var ErrSessionNotFound = errors.New("session storage: not found")

// FileSessionStorage implements SessionStorage for file system as file
// stored in Path.
type FileSessionStorage struct {
	Path string
}

// LoadSession loads session from file.
func (f *FileSessionStorage) LoadSession(_ context.Context) ([]byte, error) {
	if f == nil {
		return nil, xerrors.New("nil session storage is invalid")
	}
	data, err := ioutil.ReadFile(f.Path)
	if os.IsNotExist(err) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, xerrors.Errorf("failed to load session: %w", err)
	}
	return data, nil
}

// StoreSession stores session to file.
func (f *FileSessionStorage) StoreSession(_ context.Context, data []byte) error {
	if f == nil {
		return xerrors.New("nil session storage is invalid")
	}
	return ioutil.WriteFile(f.Path, data, 0600)
}

// NB: any changes to this structure will break backward compatibility
// to all clients.
type jsonSession struct {
	Salt      int64  `json:"salt"`
	AuthKey   []byte `json:"auth_key"`
	AuthKeyID []byte `json:"auth_key_id"`
}

func (c *Client) saveSession(ctx context.Context) error {
	if c.sessionStorage == nil {
		return nil
	}

	sess := jsonSession{
		Salt:      atomic.LoadInt64(&c.salt),
		AuthKeyID: c.authKeyID[:],
		AuthKey:   c.authKey[:],
	}
	data, err := json.Marshal(sess)
	if err != nil {
		return xerrors.Errorf("failed to marshal session: %w", err)
	}
	if err := c.sessionStorage.StoreSession(ctx, data); err != nil {
		return xerrors.Errorf("failed to store session: %w", err)
	}

	return nil
}

func (c *Client) loadSession(ctx context.Context) error {
	if c.sessionStorage == nil {
		return nil
	}
	data, err := c.sessionStorage.LoadSession(ctx)
	if errors.Is(err, ErrSessionNotFound) {
		// Will create session after key exchange.
		return nil
	}

	// NB: Any change to unmarshalling can break clients in backward
	// incompatible way.
	var sess jsonSession
	if err := json.Unmarshal(data, &sess); err != nil {
		// Probably we can re-create session anyway via explicit config
		// option.
		return xerrors.Errorf("failed to unmarshal session: %w", err)
	}

	// Validating auth key.
	var authKey crypto.AuthKey
	copy(authKey[:], sess.AuthKey)
	var authKeyID [8]byte
	copy(authKeyID[:], sess.AuthKeyID)

	if authKey.ID() != authKeyID {
		return xerrors.New("auth key id does not match (corrupted session data)")
	}

	// Success.
	c.authKey = authKey
	c.authKeyID = authKeyID
	atomic.StoreInt64(&c.salt, sess.Salt)
	c.log.Info("Session loaded from storage")

	// Generating new session id.
	sessID, err := crypto.NewSessionID(c.rand)
	if err != nil {
		return err
	}
	atomic.StoreInt64(&c.session, sessID)

	return nil
}
