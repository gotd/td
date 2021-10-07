package transport

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/internal/testutil"
)

type closeMockListener struct {
	closed int
	err    error
}

func (m *closeMockListener) Accept() (net.Conn, error) {
	panic("unexpected call")
}

func (m *closeMockListener) Addr() net.Addr {
	panic("unexpected call")
}

func (m *closeMockListener) Close() error {
	m.closed++
	return m.err
}

func Test_onceCloseListener_Close(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		m := &closeMockListener{}
		once := onceCloseListener{Listener: m}
		require.NoError(t, once.Close())
		require.NoError(t, once.Close())
		require.Equal(t, 1, m.closed)
	})

	t.Run("With Error", func(t *testing.T) {
		testErr := testutil.TestError()
		m := &closeMockListener{err: testErr}
		once := onceCloseListener{Listener: m}
		require.Equal(t, testErr, once.Close())
		require.Equal(t, testErr, once.Close())
		require.Equal(t, 1, m.closed)
	})
}
