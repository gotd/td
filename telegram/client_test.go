package telegram

import (
	"context"
	"crypto/md5"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/rpc"
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/telegram/internal/rpcmock"
	"github.com/gotd/td/tg"
)

type testHandler func(id int64, body bin.Encoder) (bin.Encoder, error)

type testConn struct {
	id     atomic.Int64
	engine *rpc.Engine
	ready  *tdsync.Ready
}

func (t *testConn) Ready() <-chan struct{} {
	return t.ready.Ready()
}

func (t *testConn) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	id := t.id.Inc() - 1
	return t.engine.Do(ctx, rpc.Request{
		Input:  input,
		Output: output,
		MsgID:  id,
	})
}

func (testConn) Run(ctx context.Context) error { return nil }

func newTestClient(h testHandler) *Client {
	var engine *rpc.Engine

	engine = rpc.New(func(ctx context.Context, msgID int64, seqNo int32, in bin.Encoder) error {
		if response, err := h(msgID, in); err != nil {
			engine.NotifyError(msgID, err)
		} else {
			var b bin.Buffer
			if err := b.Encode(response); err != nil {
				return err
			}
			return engine.NotifyResult(msgID, &b)
		}
		return nil
	}, rpc.Options{})

	ready := tdsync.NewReady()
	ready.Signal()
	client := &Client{
		log:     zap.NewNop(),
		rand:    rand.New(rand.NewSource(1)),
		appID:   TestAppID,
		appHash: TestAppHash,
		conn:    &testConn{engine: engine, ready: ready},
		ctx:     context.Background(),
		cancel:  func() {},
	}
	client.init()

	return client
}

func mockClient(cb func(mock *rpcmock.Mock, client *Client)) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		a := require.New(t)
		mock := rpcmock.NewMock(t, a)
		client := newTestClient(testHandler(mock.Handler()))
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
