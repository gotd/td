package tgtest

import (
	"context"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/assert"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/proto/codec"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/transport"
)

func Test_exchangeConn_Recv(t *testing.T) {
	a := assert.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	i := transport.Intermediate
	c1, c2 := i.Pipe()
	defer func() {
		a.NoError(c1.Close())
		a.NoError(c2.Close())
	}()
	e := exchangeConn{Conn: c1}

	s := "abcdabcd"
	a.Len(s, 8)

	grp := tdsync.NewCancellableGroup(ctx)
	grp.Go(func(ctx context.Context) error {
		b := bin.Buffer{Buf: []byte(s)}
		if err := c2.Send(ctx, &b); err != nil {
			return err
		}

		b.Reset()
		var protocolErr *codec.ProtocolErr
		if err := c2.Recv(ctx, &b); err != nil && !errors.As(err, &protocolErr) {
			return err
		}

		b.ResetN(8)
		b.Put([]byte(s))
		if err := c2.Send(ctx, &b); err != nil {
			return err
		}

		return nil
	})

	var b bin.Buffer
	a.NoError(e.Recv(ctx, &b))
	b.Skip(8)
	a.Equal(s, string(b.Buf))

	a.NoError(grp.Wait())
}
