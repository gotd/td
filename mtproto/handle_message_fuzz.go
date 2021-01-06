// +build fuzz

package mtproto

import (
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/rpc"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/tg"
)

type Zero struct{}

func (Zero) Read(p []byte) (n int, err error) { return len(p), nil }

type fuzzHandler struct {
	types *tmap.Constructor
}

func (h fuzzHandler) OnMessage(b *bin.Buffer) error {
	id, err := b.PeekID()
	if err != nil {
		return err
	}
	v := h.types.New(id)
	if v == nil {
		return xerrors.New("not found")
	}
	if err := v.Decode(b); err != nil {
		return xerrors.Errorf("decode: %w", err)
	}
	return nil
}

func (fuzzHandler) OnSession(session Session) error { return nil }

func FuzzHandleMessage(data []byte) int {
	handler := fuzzHandler{
		// Handler will try to dynamically decode any incoming message.
		types: tmap.NewConstructor(
			tg.TypesConstructorMap(),
			mt.TypesConstructorMap(),
		),
	}
	c := &Conn{
		rand:      Zero{},
		rpc:       rpc.New(rpc.NopSend, rpc.Options{}),
		log:       zap.NewNop(),
		messageID: proto.NewMessageIDGen(time.Now, 1),
		handler:   handler,
	}
	if err := c.handleMessage(&bin.Buffer{Buf: data}); err != nil {
		return 0
	}
	return 1
}
