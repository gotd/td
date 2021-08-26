package tdesktop

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrongMagicError_Error(t *testing.T) {
	w := &WrongMagicError{}
	require.NotEmpty(t, w.Error())
}
