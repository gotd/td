package mtproto

import (
	"context"
	"crypto/md5"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/internal/mt"
	"github.com/nnqq/td/internal/proto"
	"github.com/nnqq/td/internal/rpc"
	"github.com/nnqq/td/internal/tmap"
	"github.com/nnqq/td/tg"
)

type testHandler func(msgID int64, seqNo int32, body bin.Encoder) (bin.Encoder, error)

type testClientOption func(o Options)

func newTestClient(h testHandler, opts ...testClientOption) *Conn {
	var engine *rpc.Engine

	engine = rpc.New(func(ctx context.Context, msgID int64, seqNo int32, in bin.Encoder) error {
		if response, err := h(msgID, seqNo, in); err != nil {
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

	opt := Options{
		Logger:    zap.NewNop(),
		Random:    rand.New(rand.NewSource(1)),
		Key:       crypto.Key{}.WithID(),
		MessageID: proto.NewMessageIDGen(time.Now),

		engine: engine,
	}
	for _, o := range opts {
		o(opt)
	}

	return New(nil, opt)
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
