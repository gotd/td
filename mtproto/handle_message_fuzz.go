//go:build fuzz
// +build fuzz

package mtproto

import (
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/rpc"
	"github.com/gotd/td/testutil"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tmap"
)

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
		return errors.New("not found")
	}
	if err := v.Decode(b); err != nil {
		return errors.Wrap(err, "decode")
	}

	// Performing decode cycle.
	var newBuff bin.Buffer
	newV := h.types.New(id)
	if err := v.Encode(&newBuff); err != nil {
		panic(err)
	}
	if err := newV.Decode(&newBuff); err != nil {
		panic(err)
	}

	return nil
}

func (fuzzHandler) OnSession(session Session) error { return nil }

var (
	conn *Conn
	buf  *bin.Buffer
)

func init() {
	handler := fuzzHandler{
		// Handler will try to dynamically decode any incoming message.
		types: tmap.NewConstructor(
			tg.TypesConstructorMap(),
			mt.TypesConstructorMap(),
		),
	}
	c := &Conn{
		rand:      testutil.ZeroRand{},
		rpc:       rpc.New(rpc.NopSend, rpc.Options{}),
		log:       zap.NewNop(),
		messageID: proto.NewMessageIDGen(time.Now),
		handler:   handler,
	}

	conn = c
	buf = &bin.Buffer{}
}

func FuzzHandleMessage(data []byte) int {
	buf.ResetTo(data)
	if err := conn.handleMessage(buf); err != nil {
		return 0
	}
	return 1
}
