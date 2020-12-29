package telegram

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
)

func TestTrace_OnMessage(t *testing.T) {
	client := newTestClient(nil)
	var traceCalled bool
	client.trace.Message = func(b *bin.Buffer) {
		assert.Empty(t, b.Buf)
		traceCalled = true
	}
	_ = client.handleMessage(&bin.Buffer{Buf: nil})
	require.True(t, traceCalled, "trace should be called")
}
