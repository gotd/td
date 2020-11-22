package bin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFields(t *testing.T) {
	var f Fields
	f.Set(1)
	f.Set(0)
	f.Set(5)
	require.True(t, f.Has(1))
	require.True(t, f.Has(5))
	require.True(t, f.Has(0))
	require.False(t, f.Has(2))
	t.Run("Encode", func(t *testing.T) {
		var b Buffer
		require.NoError(t, f.Encode(&b))
		var decoded Fields
		require.NoError(t, decoded.Decode(&b))
		require.Equal(t, f, decoded)
	})
}
