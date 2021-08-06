package wsutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddr(t *testing.T) {
	a := require.New(t)

	s := "localhost"
	addr := Addr(s)
	a.Equal("websocket", addr.Network())
	a.Equal(s, addr.String())
}
