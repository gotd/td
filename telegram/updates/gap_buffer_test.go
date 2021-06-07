package updates

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGapBuffer(t *testing.T) {
	buf := new(gapBuffer)
	buf.Enable(1, 7)

	require.True(t, buf.Consume(update{State: 4, Count: 3}))
	require.Equal(t, []gap{{4, 7}}, buf.gaps)

	require.False(t, buf.Consume(update{State: 1, Count: 1}))
	require.Equal(t, []gap{{4, 7}}, buf.gaps)

	require.False(t, buf.Consume(update{State: 8, Count: 1}))
	require.Equal(t, []gap{{4, 7}}, buf.gaps)

	require.True(t, buf.Consume(update{State: 7, Count: 3}))
	require.Empty(t, buf.gaps)
}
