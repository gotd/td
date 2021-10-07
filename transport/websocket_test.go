package transport_test

import (
	"context"
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/telegram/dcs"
	"github.com/nnqq/td/transport"
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
			return xerrors.Errorf("accept: %w", err)
		}

		var b bin.Buffer
		if err := conn.Recv(ctx, &b); err != nil {
			return xerrors.Errorf("recv: %w", err)
		}

		if err := conn.Send(ctx, &b); err != nil {
			return xerrors.Errorf("send: %w", err)
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
