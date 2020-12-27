package mtproto

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/mtproto/internal/rpc"
)

type testHandler func(id int64, body bin.Encoder) (bin.Encoder, error)

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
		rpc:            engine,
		log:            zap.NewNop(),
		clock:          time.Now,
		rand:           rand.New(rand.NewSource(1)),
		sessionCreated: createCondOnce(),
		appID:          TestAppID,
		appHash:        TestAppHash,
		authKey:        crypto.AuthKey{}.WithID(),
	}
	client.sessionCreated.Done()

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
