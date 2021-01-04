package telegram

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gotd/td/mtproto"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/rpc"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/tg"
)

type testHandler func(id int64, body bin.Encoder) (bin.Encoder, error)

func testError(err tg.Error) (bin.Encoder, error) {
	e := &mtproto.Error{
		Message: err.Text,
		Code:    err.Code,
	}
	e.ExtractArgument()
	return nil, e
}

type testConn struct {
	id     int64
	engine *rpc.Engine
}

func (t *testConn) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	t.id++
	return t.engine.Do(ctx, rpc.Request{
		Input:    input,
		Output:   output,
		ID:       t.id,
		Sequence: int32(t.id),
	})
}

func (testConn) Connect(ctx context.Context) error { return nil }
func (testConn) Config() tg.Config                 { return tg.Config{} }
func (testConn) Close() error                      { return nil }

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
	}, rpc.Config{})

	client := &Client{
		log:     zap.NewNop(),
		rand:    rand.New(rand.NewSource(1)),
		appID:   TestAppID,
		appHash: TestAppHash,
		conn:    &testConn{engine: engine},
	}
	client.tg = tg.NewClient(client)

	return client
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
		assert.NoError(t, ioutil.WriteFile(filepath.Join(dir, base), b.Buf, 0600))
	}
}
