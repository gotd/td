package mtproto

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/crypto"
)

func TestNewPFSUsesPermKey(t *testing.T) {
	a := require.New(t)
	perm := crypto.Key{7}.WithID()

	conn := New(nil, Options{
		EnablePFS: true,
		PermKey:   perm,
	})
	session := conn.session()

	// Runtime key must start empty in PFS mode until temporary key is generated.
	a.True(session.Key.Zero())
	a.Equal(perm, session.PermKey)
}

func TestTempKeyRenewalLoopReconnect(t *testing.T) {
	a := require.New(t)

	conn := Conn{
		pfs:           true,
		tempKeyTTL:    60,
		tempKeyExpiry: time.Now().Add(-time.Second).Unix(),
		clock:         clock.System,
		log:           zap.NewNop(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := conn.tempKeyRenewalLoop(ctx)
	a.Error(err)
	a.ErrorIs(err, ErrPFSReconnectRequired)
}

func TestHandleAuthKeyNotFoundPFS(t *testing.T) {
	a := require.New(t)
	conn := New(nil, Options{
		EnablePFS: true,
		Logger:    zap.NewNop(),
	})

	err := conn.handleAuthKeyNotFound(context.Background())
	a.Error(err)
	a.ErrorIs(err, ErrPFSReconnectRequired)
}

func TestPermKeyOlderThan60s(t *testing.T) {
	a := require.New(t)
	now := time.Now().Unix()
	conn := Conn{
		clock: clock.System,
	}

	// Quote (PFS): "if ... ENCRYPTED_MESSAGE_INVALID and permanent key was generated more than 60 seconds ago..."
	// Link: https://core.telegram.org/api/pfs
	conn.permKeyCreatedAt = 0
	a.False(conn.permKeyAgeKnown())
	a.False(conn.permKeyOlderThan60s())

	conn.permKeyCreatedAt = now - 61
	a.True(conn.permKeyAgeKnown())
	a.True(conn.permKeyOlderThan60s())

	conn.permKeyCreatedAt = now - 10
	a.False(conn.permKeyOlderThan60s())
}

func TestHandleBindEncryptedMessageInvalidDropsOldKeys(t *testing.T) {
	a := require.New(t)
	conn := Conn{
		clock:            clock.System,
		log:              zap.NewNop(),
		authKey:          crypto.Key{1}.WithID(),
		permKey:          crypto.Key{2}.WithID(),
		permKeyCreatedAt: time.Now().Unix() - 61,
		salt:             10,
		sessionID:        11,
		tempKeyExpiry:    12,
	}

	err := conn.handleBindEncryptedMessageInvalid(1)
	a.Error(err)
	a.ErrorIs(err, ErrPFSDropKeysRequired)

	// Both keys and session data are reset so reconnect starts from clean state.
	session := conn.session()
	a.True(session.Key.Zero())
	a.True(session.PermKey.Zero())
	a.Zero(session.Salt)
	a.Zero(session.ID)
}

func TestHandleBindEncryptedMessageInvalidRetryFreshKey(t *testing.T) {
	a := require.New(t)
	temp := crypto.Key{3}.WithID()
	perm := crypto.Key{4}.WithID()
	conn := Conn{
		clock:            clock.System,
		log:              zap.NewNop(),
		authKey:          temp,
		permKey:          perm,
		permKeyCreatedAt: time.Now().Unix(),
		salt:             20,
		sessionID:        21,
		tempKeyExpiry:    22,
	}

	err := conn.handleBindEncryptedMessageInvalid(1)
	a.NoError(err)

	session := conn.session()
	a.Equal(temp, session.Key)
	a.Equal(perm, session.PermKey)
	a.Equal(int64(20), session.Salt)
	a.Equal(int64(21), session.ID)
}

func TestHandleBindEncryptedMessageInvalidUnknownAge(t *testing.T) {
	a := require.New(t)
	temp := crypto.Key{7}.WithID()
	perm := crypto.Key{8}.WithID()
	conn := Conn{
		clock:         clock.System,
		log:           zap.NewNop(),
		authKey:       temp,
		permKey:       perm,
		salt:          30,
		sessionID:     31,
		tempKeyExpiry: 32,
	}

	// Unknown key age should retry first.
	err := conn.handleBindEncryptedMessageInvalid(1)
	a.NoError(err)
	session := conn.session()
	a.Equal(temp, session.Key)
	a.Equal(perm, session.PermKey)
	a.Equal(int64(30), session.Salt)
	a.Equal(int64(31), session.ID)

	// After retries are exhausted, key reset is forced to avoid reconnect loop.
	err = conn.handleBindEncryptedMessageInvalid(maxBindTempAuthKeyAttempts)
	a.Error(err)
	a.ErrorIs(err, ErrPFSDropKeysRequired)

	session = conn.session()
	a.True(session.Key.Zero())
	a.True(session.PermKey.Zero())
	a.Zero(session.Salt)
	a.Zero(session.ID)
}

func TestHandleBindConnectionNotInitedRetriesThenReconnects(t *testing.T) {
	a := require.New(t)
	temp := crypto.Key{5}.WithID()
	perm := crypto.Key{6}.WithID()
	conn := Conn{
		clock:         clock.System,
		log:           zap.NewNop(),
		authKey:       temp,
		permKey:       perm,
		salt:          30,
		sessionID:     31,
		tempKeyExpiry: 32,
	}

	err := conn.handleBindConnectionNotInited(1)
	a.NoError(err)

	session := conn.session()
	a.Equal(temp, session.Key)
	a.Equal(perm, session.PermKey)
	a.Equal(int64(30), session.Salt)
	a.Equal(int64(31), session.ID)

	err = conn.handleBindConnectionNotInited(maxBindTempAuthKeyAttempts)
	a.Error(err)
	// CONNECTION_NOT_INITED path should not purge persisted permanent key.
	a.NotErrorIs(err, ErrPFSDropKeysRequired)

	session = conn.session()
	a.Equal(temp, session.Key)
	a.Equal(perm, session.PermKey)
	a.Equal(int64(30), session.Salt)
	a.Equal(int64(31), session.ID)
}
