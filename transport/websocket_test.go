package transport_test

import (
	"context"
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

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

	listener, h := transport.WebsocketListener(srv.URL)
	handler = h
	list := dcs.List{
		Domains: map[int]string{
			2: srv.URL,
		},
	}

	server := transport.NewCustomServer(nil, listener)
	defer server.Close()
	done := make(chan struct{})

	go func() {
		close(done)
		_ = server.Serve(ctx, func(ctx context.Context, conn transport.Conn) error {
			var b bin.Buffer
			for {
				b.Reset()

				if err := conn.Recv(ctx, &b); err != nil {
					return xerrors.Errorf("recv: %w", err)
				}

				if err := conn.Send(ctx, &b); err != nil {
					return xerrors.Errorf("send: %w", err)
				}
			}
		})
	}()

	rs := dcs.Websocket(dcs.WebsocketOptions{})
	conn, err := rs.Primary(ctx, 2, list)
	a.NoError(err)

	data, err := io.ReadAll(io.LimitReader(rand.Reader, 1024))
	a.NoError(err)
	a.NoError(conn.Send(ctx, &bin.Buffer{Buf: data}))

	var b bin.Buffer
	a.NoError(conn.Recv(ctx, &b))
	a.Equal(data, b.Buf)

	<-done
}
