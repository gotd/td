package dcs

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
)

// The poll loop must fail over past a dead candidate URL to a working one.
func TestHTTPConn_FailoverRotatesPastDeadURL(t *testing.T) {
	a := require.New(t)

	// A closed server yields a guaranteed connection-refused address.
	deadSrv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	deadURL := deadSrv.URL + "/api"
	deadSrv.Close()

	payload := []byte("alive")
	alive := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		_, _ = w.Write(payload)
	}))
	defer alive.Close()

	client := &http.Client{Timeout: 2 * time.Second}
	// Dead address first: the loop must rotate to the working one.
	conn := newHTTPConn(client, []string{deadURL, alive.URL + "/api"}, 0, 0, defaultHTTPMaxWait)
	defer func() { _ = conn.Close() }()

	conn.StartHTTPWait(func(context.Context) (*bin.Buffer, error) {
		return &bin.Buffer{Buf: []byte("wait")}, nil
	})

	recvCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var b bin.Buffer
	a.NoError(conn.Recv(recvCtx, &b))
	a.Equal(payload, b.Buf)
}
