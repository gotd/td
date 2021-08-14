package bin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFields(t *testing.T) {
	var f Fields
	a := require.New(t)

	a.True(f.Zero())
	f.Set(1)
	f.Set(0)
	f.Set(10)
	f.Unset(10)
	f.Set(5)
	a.True(f.Has(1))
	a.True(f.Has(5))
	a.True(f.Has(0))
	a.False(f.Has(2))
	a.False(f.Has(10))
	a.Equal("100011", f.String())

	t.Run("Encode", func(t *testing.T) {
		a := require.New(t)

		var b Buffer
		a.NoError(f.Encode(&b))
		var decoded Fields
		a.NoError(decoded.Decode(&b))
		a.Equal(f, decoded)
	})

	t.Run("Decode", func(t *testing.T) {
		var decoded Fields
		a.Error(decoded.Decode(&Buffer{}))
	})
}
