package rpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNopDrop(t *testing.T) {
	require.NoError(t, NopDrop(Request{}))
}

func TestNopSend(t *testing.T) {
	require.NoError(t, NopSend(context.TODO(), 0, 0, nil))
}
