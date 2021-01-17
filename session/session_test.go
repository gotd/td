package session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientSessionStorage(t *testing.T) {
	loader := Loader{
		Storage: &StorageMemory{},
	}
	ctx := context.Background()
	t.Run("Ok", func(t *testing.T) {
		{
			_, err := loader.Load(ctx)
			require.ErrorIs(t, err, ErrNotFound)
		}
		data := &Data{DC: 2}
		require.NoError(t, loader.Save(ctx, data))
		{
			gotData, err := loader.Load(ctx)
			require.NoError(t, err)
			require.Equal(t, data, gotData)
		}
	})
}
