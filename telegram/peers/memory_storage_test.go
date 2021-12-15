package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemoryStorage(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)
	k := Key{
		Prefix: usersPrefix,
		ID:     1,
	}
	v := Value{
		AccessHash: 10,
	}
	phone := "phone"

	var m InmemoryStorage
	_, found, err := m.Find(ctx, k)
	a.NoError(err)
	a.False(found)

	a.NoError(m.Save(ctx, k, v))

	v2, found, err := m.Find(ctx, k)
	a.NoError(err)
	a.True(found)
	a.Equal(v, v2)

	_, _, found, err = m.FindPhone(ctx, phone)
	a.NoError(err)
	a.False(found)

	a.NoError(m.SavePhone(ctx, phone, k))

	k2, v2, found, err := m.FindPhone(ctx, phone)
	a.NoError(err)
	a.True(found)
	a.Equal(k, k2)
	a.Equal(v, v2)

	hash, err := m.GetContactsHash(ctx)
	a.NoError(err)
	a.Zero(hash)

	a.NoError(m.SaveContactsHash(ctx, 1))

	hash, err = m.GetContactsHash(ctx)
	a.NoError(err)
	a.Equal(int64(1), hash)
}
