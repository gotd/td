package telegram

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/session"
)

func TestRestoreConnection_DCOnly(t *testing.T) {
	ctx := context.Background()
	st := &session.StorageMemory{}
	loader := &session.Loader{Storage: st}
	require.NoError(t, loader.Save(ctx, &session.Data{DC: 4}))

	c := NewClient(1, "hash", Options{
		SessionStorage: st,
		DC:             2,
		NoUpdates:      true,
	})
	require.NoError(t, c.restoreConnection(ctx))

	got := c.session.Load()
	require.Equal(t, 4, got.DC)
	require.True(t, got.AuthKey.Zero())
}

func TestRestoreConnection_MissingAuthKeyID(t *testing.T) {
	ctx := context.Background()
	var k crypto.Key
	_, err := rand.Read(k[:])
	require.NoError(t, err)
	ak := k.WithID()

	st := &session.StorageMemory{}
	loader := &session.Loader{Storage: st}
	require.NoError(t, loader.Save(ctx, &session.Data{
		DC:      3,
		AuthKey: ak.Value[:],
	}))

	c := NewClient(1, "hash", Options{
		SessionStorage: st,
		DC:             2,
		NoUpdates:      true,
	})
	require.NoError(t, c.restoreConnection(ctx))

	got := c.session.Load()
	require.Equal(t, 3, got.DC)
	require.Equal(t, ak.ID, got.AuthKey.ID)
	require.Equal(t, ak.Value, got.AuthKey.Value)
}

func TestRestoreConnection_CorruptedKey(t *testing.T) {
	ctx := context.Background()
	var k crypto.Key
	_, err := rand.Read(k[:])
	require.NoError(t, err)
	ak := k.WithID()
	wrongID := ak.ID
	wrongID[0] ^= 0xff

	st := &session.StorageMemory{}
	loader := &session.Loader{Storage: st}
	require.NoError(t, loader.Save(ctx, &session.Data{
		DC:        3,
		AuthKey:   ak.Value[:],
		AuthKeyID: wrongID[:],
	}))

	c := NewClient(1, "hash", Options{
		SessionStorage: st,
		DC:             2,
		NoUpdates:      true,
	})
	err = c.restoreConnection(ctx)
	require.EqualError(t, err, "corrupted key")
}
