package wsutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWsConn_LocalAddr(t *testing.T) {
	s := Addr("local")
	conn := NetConn(nil, s, Addr(""))
	require.Equal(t, s, conn.LocalAddr().String())
}

func TestWsConn_RemoteAddr(t *testing.T) {
	s := Addr("remote")
	conn := NetConn(nil, Addr(""), s)
	require.Equal(t, s, conn.RemoteAddr().String())
}
