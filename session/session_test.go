package session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/internal/testutil"
)

func TestClientSessionStorage(t *testing.T) {
	loader := Loader{
		Storage: &StorageMemory{},
	}
	ctx := context.Background()
	t.Run("Ok", func(t *testing.T) {
		{
			_, err := loader.Load(ctx)
			testutil.RequireErr(t, ErrNotFound, err)
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
