package transport

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/internal/testutil"
)

type mockListener struct {
	closed int
	err    error
}

func (m *mockListener) Accept() (net.Conn, error) {
	panic("unexpected call")
}

func (m *mockListener) Addr() net.Addr {
	panic("unexpected call")
}

func (m *mockListener) Close() error {
	m.closed++
	return m.err
}

func Test_onceCloseListener_Close(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		m := &mockListener{}
		once := onceCloseListener{Listener: m}
		require.NoError(t, once.Close())
		require.NoError(t, once.Close())
		require.Equal(t, 1, m.closed)
	})

	t.Run("With Error", func(t *testing.T) {
		testErr := testutil.TestError()
		m := &mockListener{err: testErr}
		once := onceCloseListener{Listener: m}
		require.Equal(t, testErr, once.Close())
		require.Equal(t, testErr, once.Close())
		require.Equal(t, 1, m.closed)
	})
}
