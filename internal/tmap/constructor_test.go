package tmap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConstructor_New(t *testing.T) {
	require.Nil(t, (&Constructor{}).New(0))
}
