package telegram

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorParse(t *testing.T) {
	e := Error{
		Code:    420,
		Message: "FLOOD_WAIT_359",
	}
	e.extractArgument()

	require.Equal(t, Error{
		Code:     420,
		Message:  "FLOOD_WAIT_359",
		Type:     "FLOOD_WAIT",
		Argument: 359,
	}, e)
}
