package mtproto

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/log"
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
		log:           log.For(log.Nop),
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
		Logger:    log.Nop,
	})

	err := conn.handleAuthKeyNotFound(context.Background())
	a.Error(err)
	a.ErrorIs(err, ErrPFSReconnectRequired)
}

func stubPermKeyCheck(unknown bool, calls *int) func(context.Context) (bool, error) {
	return func(context.Context) (bool, error) {
		*calls++
		return unknown, nil
	}
}

func TestHandleBindEncryptedMessageInvalidDropsUnknownKey(t *testing.T) {
	a := require.New(t)
	calls := 0
	conn := Conn{
		clock:            clock.System,
		log:              log.For(log.Nop),
		authKey:          crypto.Key{1}.WithID(),
		permKey:          crypto.Key{2}.WithID(),
		permKeyCreatedAt: time.Now().Unix() - 61,
		salt:             10,
		sessionID:        11,
		tempKeyExpiry:    12,
		permKeyCheck:     stubPermKeyCheck(true, &calls),
	}

	err := conn.handleBindEncryptedMessageInvalid(context.Background(), 1)
	a.ErrorIs(err, ErrPFSDropKeysRequired)
	a.Equal(1, calls)

	// Both keys and session data are reset so reconnect starts from clean state.
	session := conn.session()
	a.True(session.Key.Zero())
	a.True(session.PermKey.Zero())
	a.Zero(session.Salt)
	a.Zero(session.ID)
}

// An old key the server still knows is kept: retry, then reconnect.
func TestHandleBindEncryptedMessageInvalidKeepsKnownKey(t *testing.T) {
	a := require.New(t)
	calls := 0
	perm := crypto.Key{2}.WithID()
	conn := Conn{
		clock:            clock.System,
		log:              log.For(log.Nop),
		authKey:          crypto.Key{1}.WithID(),
		permKey:          perm,
		permKeyCreatedAt: time.Now().Unix() - 3600,
		permKeyCheck:     stubPermKeyCheck(false, &calls),
	}
	a.NoError(conn.handleBindEncryptedMessageInvalid(context.Background(), 1))
	err := conn.handleBindEncryptedMessageInvalid(context.Background(), maxBindTempAuthKeyAttempts)
	a.ErrorIs(err, ErrPFSReconnectRequired)
	a.NotErrorIs(err, ErrPFSDropKeysRequired)
	a.Equal(perm, conn.session().PermKey)
	a.Equal(2, calls)
}

func TestHandleBindEncryptedMessageInvalidRetryFreshKey(t *testing.T) {
	a := require.New(t)
	temp := crypto.Key{3}.WithID()
	perm := crypto.Key{4}.WithID()
	calls := 0
	conn := Conn{
		clock:            clock.System,
		log:              log.For(log.Nop),
		authKey:          temp,
		permKey:          perm,
		permKeyCreatedAt: time.Now().Unix(),
		salt:             20,
		sessionID:        21,
		tempKeyExpiry:    22,
		permKeyCheck:     stubPermKeyCheck(true, &calls),
	}

	a.NoError(conn.handleBindEncryptedMessageInvalid(context.Background(), 1))
	a.Zero(calls, "a fresh key is not checked")

	session := conn.session()
	a.Equal(temp, session.Key)
	a.Equal(perm, session.PermKey)
	a.Equal(int64(20), session.Salt)
	a.Equal(int64(21), session.ID)

	// After retries: reconnect, keep the key.
	err := conn.handleBindEncryptedMessageInvalid(context.Background(), maxBindTempAuthKeyAttempts)
	a.ErrorIs(err, ErrPFSReconnectRequired)
	a.Equal(perm, conn.session().PermKey)
}

// A key older than a day that was bound within the last day is immune.
func TestHandleBindEncryptedMessageInvalidImmuneOldKey(t *testing.T) {
	a := require.New(t)
	calls := 0
	perm := crypto.Key{4}.WithID()
	state := NewSessionState()
	state.setLastBind(time.Now().Unix() - 3600)
	conn := Conn{
		clock:            clock.System,
		log:              log.For(log.Nop),
		authKey:          crypto.Key{3}.WithID(),
		permKey:          perm,
		permKeyCreatedAt: time.Now().Unix() - 2*86400,
		sessionState:     state,
		permKeyCheck:     stubPermKeyCheck(true, &calls),
	}
	a.NoError(conn.handleBindEncryptedMessageInvalid(context.Background(), 1))
	a.Zero(calls)
	a.Equal(perm, conn.session().PermKey)
}

// A restored key of unknown age is checked, not dropped blindly.
func TestHandleBindEncryptedMessageInvalidUnknownAge(t *testing.T) {
	a := require.New(t)
	temp := crypto.Key{7}.WithID()
	perm := crypto.Key{8}.WithID()
	calls := 0
	conn := Conn{
		clock:         clock.System,
		log:           log.For(log.Nop),
		authKey:       temp,
		permKey:       perm,
		salt:          30,
		sessionID:     31,
		tempKeyExpiry: 32,
		permKeyCheck:  stubPermKeyCheck(false, &calls),
	}

	a.NoError(conn.handleBindEncryptedMessageInvalid(context.Background(), 1))
	a.Equal(1, calls)
	session := conn.session()
	a.Equal(temp, session.Key)
	a.Equal(perm, session.PermKey)

	err := conn.handleBindEncryptedMessageInvalid(context.Background(), maxBindTempAuthKeyAttempts)
	a.ErrorIs(err, ErrPFSReconnectRequired)
	a.Equal(perm, conn.session().PermKey, "known key is not dropped")
}

func TestBindRetryDelay(t *testing.T) {
	require.Equal(t, []time.Duration{time.Second, 2 * time.Second, 4 * time.Second},
		[]time.Duration{bindRetryDelay(1), bindRetryDelay(2), bindRetryDelay(3)})
}

func TestHandleBindConnectionNotInitedRetriesThenReconnects(t *testing.T) {
	a := require.New(t)
	temp := crypto.Key{5}.WithID()
	perm := crypto.Key{6}.WithID()
	conn := Conn{
		clock:         clock.System,
		log:           log.For(log.Nop),
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
