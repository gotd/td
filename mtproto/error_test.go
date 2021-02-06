package mtproto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorParse(t *testing.T) {
	t.Run("FLOOD_WAIT", func(t *testing.T) {
		require.Equal(t, &Error{
			Code:     420,
			Message:  "FLOOD_WAIT_359",
			Type:     "FLOOD_WAIT",
			Argument: 359,
		}, NewError(420, "FLOOD_WAIT_359"))
	})
	t.Run("Middle", func(t *testing.T) {
		require.Equal(t, &Error{
			Code:     169,
			Message:  "GO_1337_METERS_AWAY",
			Type:     "GO_METERS_AWAY",
			Argument: 1337,
		}, NewError(169, "GO_1337_METERS_AWAY"))
	})
}

func TestHelpers(t *testing.T) {
	err := func() error {
		return NewError(169, "GO_1337_METERS_AWAY")
	}()
	t.Run("Type", func(t *testing.T) {
		assert.True(t, IsErr(err, "GO_METERS_AWAY"))
		assert.True(t, IsErr(err, "FOO", "GO_METERS_AWAY"))
		assert.False(t, IsErr(err, "NOPE"))
		t.Run("AsType", func(t *testing.T) {
			{
				rpcErr, ok := AsTypeErr(err, "NOPE")
				require.False(t, ok)
				require.Nil(t, rpcErr)
			}
			{
				rpcErr, ok := AsTypeErr(err, "GO_METERS_AWAY")
				require.True(t, ok)
				require.NotNil(t, rpcErr)
			}
		})
	})
	t.Run("Code", func(t *testing.T) {
		assert.True(t, IsErrCode(err, 169))
		assert.True(t, IsErrCode(err, 1, 169))
		assert.False(t, IsErrCode(err, 168))
	})
}
