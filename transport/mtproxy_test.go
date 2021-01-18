package transport

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMTProxy(t *testing.T) {
	_, err := MTProxy(nil, 0, nil)
	require.Error(t, err)
}
