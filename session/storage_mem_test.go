package session

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorageMemory(t *testing.T) {
	t.Run("Storage", testStorage(&StorageMemory{}))
}

func TestStorageMemory_Dump(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	data := bytes.Repeat([]byte{'a'}, 256)

	a.ErrorIs((*StorageMemory)(nil).Dump(nil), ErrNotFound)

	s := StorageMemory{}
	a.ErrorIs(s.Dump(nil), ErrNotFound)

	a.NoError(s.StoreSession(ctx, data))
	out := bytes.Buffer{}
	a.NoError(s.Dump(&out))
	a.Equal(data, out.Bytes())
}

func TestStorageMemory_Clone(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	data := bytes.Repeat([]byte{'a'}, 256)
	tmp := data[0]

	s := StorageMemory{}
	a.NoError(s.StoreSession(ctx, data))

	s2 := s.Clone()
	s.data[0]--
	a.Equal(tmp, s2.data[0])
}
