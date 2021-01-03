package mtproto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorParse(t *testing.T) {
	t.Run("FLOOD_WAIT", func(t *testing.T) {
		e := Error{
			Code:    420,
			Message: "FLOOD_WAIT_359",
		}
		e.ExtractArgument()

		require.Equal(t, Error{
			Code:     420,
			Message:  "FLOOD_WAIT_359",
			Type:     "FLOOD_WAIT",
			Argument: 359,
		}, e)
	})
	t.Run("Middle", func(t *testing.T) {
		e := Error{
			Code:    169,
			Message: "GO_1337_METERS_AWAY",
		}
		e.ExtractArgument()

		require.Equal(t, Error{
			Code:     169,
			Message:  "GO_1337_METERS_AWAY",
			Type:     "GO_METERS_AWAY",
			Argument: 1337,
		}, e)
	})
}
