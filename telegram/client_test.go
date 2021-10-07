package telegram

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/mt"
	"github.com/nnqq/td/internal/proto"
	"github.com/nnqq/td/internal/rpc"
	"github.com/nnqq/td/internal/tdsync"
	"github.com/nnqq/td/internal/testutil"
	"github.com/nnqq/td/internal/tmap"
	"github.com/nnqq/td/telegram/dcs"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgmock"
)

type testHandler func(id int64, body bin.Encoder) (bin.Encoder, error)

type testConn struct {
	id     atomic.Int64
	engine *rpc.Engine
	ready  *tdsync.Ready
}

func (t *testConn) Ping(ctx context.Context) error {
	return errors.New("not implemented")
}

func (t *testConn) Ready() <-chan struct{} {
	return t.ready.Ready()
}

func (t *testConn) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
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
		log:           zap.NewNop(),
		rand:          rand.New(rand.NewSource(1)),
		appID:         TestAppID,
		appHash:       TestAppHash,
		conn:          &testConn{engine: engine, ready: ready},
		ctx:           context.Background(),
		cancel:        func() {},
		updateHandler: UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error { return nil }),
	}
	client.init()

	return client
}

func mockClient(cb func(mock *tgmock.Mock, client *Client)) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		mock := tgmock.NewRequire(t)
		client := newTestClient(testHandler(mock.Handler()))
		cb(mock, client)
	}
}

func TestEnsureErrorIfCantConnect(t *testing.T) {
	testErr := testutil.TestError()
	dialer := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return nil, testErr
	}
	opts := Options{
		Resolver: dcs.Plain(dcs.PlainOptions{Dial: dialer}),
		ReconnectionBackoff: func() backoff.BackOff {
			return backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Nanosecond), 2)
		},
	}

	err := NewClient(1, "hash", opts).Run(context.Background(),
		func(ctx context.Context) error {
			return nil
		})
	require.ErrorIs(t, err, testErr)
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
