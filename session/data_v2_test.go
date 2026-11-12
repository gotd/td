package session

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/session/tdesktop"
)

func TestDataRoundTripNewFields(t *testing.T) {
	ctx := context.Background()
	loader := Loader{Storage: &StorageMemory{}}
	data := &Data{
		DC:               2,
		AuthKey:          bytes.Repeat([]byte{'a'}, 256),
		AuthKeyID:        []byte("gotd1337"),
		Salt:             10,
		AuthKeyCreatedAt: 1700000000,
		DCs: map[int]DCKey{
			4: {
				AuthKey:    bytes.Repeat([]byte{'b'}, 256),
				AuthKeyID:  []byte("gotd4444"),
				CreatedAt:  1700000100,
				Authorized: true,
				Salt:       5,
			},
		},
		ServerTimeOffset: 3 * time.Second,
		UserID:           10,
		Bot:              true,
	}
	require.NoError(t, loader.Save(ctx, data))
	got, err := loader.Load(ctx)
	require.NoError(t, err)
	require.Equal(t, data, got)
}

// Files written before the new fields load with them unset, and files
// without the new fields don't carry them in JSON.
func TestDataOldFormat(t *testing.T) {
	ctx := context.Background()
	old := []byte(`{"Version":1,"Data":{"Config":{"BlockedMode":false,"ForceTryIpv6":false,"Date":0,"Expires":0,"TestMode":false,"ThisDC":2,"DCOptions":null,"DCTxtDomainName":"","WebfileDCID":0,"TmpSessions":0},"DC":2,"Addr":"","AuthKey":"YWFh","AuthKeyID":"Z290ZDEzMzc=","Salt":10}}`)
	storage := &StorageMemory{}
	require.NoError(t, storage.StoreSession(ctx, old))
	got, err := (&Loader{Storage: storage}).Load(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, got.DC)
	require.Equal(t, int64(10), got.Salt)
	require.Nil(t, got.DCs)
	require.Zero(t, got.UserID)

	// Saving without new fields keeps the old shape.
	out := &StorageMemory{}
	require.NoError(t, (&Loader{Storage: out}).Save(ctx, got))
	buf, err := out.LoadSession(ctx)
	require.NoError(t, err)
	for _, field := range []string{"DCs", "UserID", "Bot", "ServerTimeOffset", "AuthKeyCreatedAt"} {
		require.NotContains(t, string(buf), field)
	}
}

func TestDataNewerVersion(t *testing.T) {
	ctx := context.Background()
	storage := &StorageMemory{}
	require.NoError(t, storage.StoreSession(ctx, []byte(`{"Version":2,"Data":{"DC":2}}`)))
	_, err := (&Loader{Storage: storage}).Load(ctx)
	require.ErrorIs(t, err, ErrVersionUnsupported)
	require.NotErrorIs(t, err, ErrNotFound, "a newer file must not look like no session")
}

func TestTDesktopSessionAllDCKeys(t *testing.T) {
	var k2, k4 crypto.Key
	k2[0], k4[0] = 2, 4
	got, err := TDesktopSession(tdesktop.Account{
		Authorization: tdesktop.MTPAuthorization{
			UserID: 10,
			MainDC: 2,
			Keys:   map[int]crypto.Key{2: k2, 4: k4},
		},
	})
	require.NoError(t, err)
	require.Equal(t, int64(10), got.UserID)
	id4 := k4.ID()
	require.Equal(t, map[int]DCKey{4: {AuthKey: k4[:], AuthKeyID: id4[:], Authorized: true}}, got.DCs)
}
