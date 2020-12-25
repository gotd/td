package tmap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMap_Get(t *testing.T) {
	require.Empty(t, (&Map{}).Get(0))
}
