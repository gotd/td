package updates

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGapBuffer(t *testing.T) {
	buf := new(gapBuffer)
	buf.Enable(1, 15)

	require.False(t, buf.Consume(update{State: 4, Count: 3}))
	require.Equal(t, []gap{{1, 1}, {5, 15}}, buf.gaps)
	require.False(t, buf.Consume(update{State: 1, Count: 1}))
	require.Equal(t, []gap{{5, 15}}, buf.gaps)
	require.True(t, buf.Consume(update{State: 15, Count: 11}))
	require.Empty(t, buf.gaps)
}
