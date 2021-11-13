package transport_test

import (
	"context"
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/transport"
)

func TestWebsocketListener(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()

	var handler http.Handler
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}))
	defer srv.Close()

	listener, h := transport.WebsocketListener(srv.Listener.Addr())
	handler = h
	list := dcs.List{
		Domains: map[int]string{
			2: srv.URL,
		},
	}

	server := transport.Listen(listener)
	defer server.Close()
	done := make(chan struct{})

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		defer close(done)

		conn, err := server.Accept()
		if err != nil {
			return errors.Wrap(err, "accept")
		}

		var b bin.Buffer
		if err := conn.Recv(ctx, &b); err != nil {
			return errors.Wrap(err, "recv")
		}

		if err := conn.Send(ctx, &b); err != nil {
			return errors.Wrap(err, "send")
		}

		return nil
	})

	rs := dcs.Websocket(dcs.WebsocketOptions{})
	conn, err := rs.Primary(ctx, 2, list)
	a.NoError(err)

	data, err := io.ReadAll(io.LimitReader(rand.Reader, 1024))
	a.NoError(err)
	a.NoError(conn.Send(ctx, &bin.Buffer{Buf: data}))

	var b bin.Buffer
	a.NoError(conn.Recv(ctx, &b))
	a.Equal(data, b.Buf)

	a.NoError(grp.Wait())
}
