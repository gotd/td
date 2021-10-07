package transport

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
)

func TestProtocol_Pipe(t *testing.T) {
	a := require.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payload := []byte("abcdabcd")
	test := func(c1, c2 Conn) {
		go func() {
			b1 := &bin.Buffer{Buf: payload}
			a.NoError(c1.Send(ctx, b1))
		}()

		b2 := &bin.Buffer{}
		a.NoError(c2.Recv(ctx, b2))
		a.Equal(payload, b2.Buf)
	}

	c1, c2 := Intermediate.Pipe()
	defer func() {
		a.NoError(c1.Close())
		a.NoError(c2.Close())
	}()

	test(c1, c2)
	test(c2, c1)
}
