package telegram

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/telegram/internal/rpcmock"
	"github.com/gotd/td/tg"
)

type testConn struct {
	handle func(body bin.Encoder) (bin.Encoder, error)
}

func (t *testConn) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	out, err := t.handle(input)
	if err != nil {
		return err
	}

	b := new(bin.Buffer)
	if err := out.Encode(b); err != nil {
		return err
	}

	return output.Decode(b)
}

func (t *testConn) Run(ctx context.Context, f func(context.Context) error) error {
	return f(ctx)
	// return nil
}

func newTestClient(h func(body bin.Encoder) (bin.Encoder, error)) *Client {
	conn := &testConn{h}
	client := &Client{
		log:     zap.NewNop(),
		appID:   TestAppID,
		appHash: TestAppHash,
		primary: conn,
	}

	return client
}

func mockClient(cb func(mock *rpcmock.Mock, client *Client)) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		a := require.New(t)
		mock := rpcmock.NewMock(t, a)
		client := newTestClient(mock.Handler())
		cb(mock, client)
	}
}

// newCorpusTracer will save incoming messages to corpus folder.
//
// Usage:
//
//	client.trace.OnMessage = newCorpusTracer(t)
//
// nolint: deadcode,unused // optional
func newCorpusTracer(t testing.TB) func(b *bin.Buffer) {
	types := tmap.New(
		mt.TypesMap(),
		tg.TypesMap(),
		proto.TypesMap(),
	)
	dir := filepath.Join("..", "_fuzz", "handle_message", "corpus")

	return func(b *bin.Buffer) {
		id, _ := b.PeekID()
		h := md5.Sum(b.Buf)
		name := types.Get(id)
		if name == "" {
			name = "unknown"
		}
		if idx := strings.Index(name, "#"); idx > 0 {
			// Removing type id from name.
			name = name[:idx]
		}
		base := fmt.Sprintf("trace_%x_%s_%x",
			id, name, h,
		)
		assert.NoError(t, os.WriteFile(filepath.Join(dir, base), b.Buf, 0600))
	}
}
