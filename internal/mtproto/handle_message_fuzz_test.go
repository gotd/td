// +build go1.17

package mtproto

import (
	"testing"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/rpc"
	"github.com/gotd/td/internal/testutil"
	"github.com/gotd/td/internal/tmap"
	"github.com/gotd/td/tg"
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
		return xerrors.New("not found")
	}
	if err := v.Decode(b); err != nil {
		return xerrors.Errorf("decode: %w", err)
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

func FuzzHandleMessage(f *testing.F) {
	types := tmap.New(
		tg.TypesMap(),
		mt.TypesMap(),
	)
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
		messageID: proto.NewMessageIDGen(time.Now),
		handler:   handler,
	}

	b := &bin.Buffer{}

	f.Fuzz(func(t *testing.T, data []byte) {
		b.ResetTo(data)
		// Default to 128 bytes per invocation.
		allocThreshold := 128

		// Adjusting threshold for specific types.
		//
		// Probably there should be better way to do this, but
		// manually ensuring allocation distribution by type is
		// pretty ok.
		b.ResetTo(data)
		if id, err := b.PeekID(); err == nil {
			t.Logf("Type: 0x%x %s", id, types.Get(id))
			switch id {
			case tg.UpdatesTypeID,
				tg.TextFixedTypeID,
				tg.InputPeerChannelFromMessageTypeID,
				tg.PageBlockRelatedArticlesTypeID:
				allocThreshold = 512
			case tg.TextBoldTypeID,
				tg.TextItalicTypeID,
				tg.TextMarkedTypeID,
				tg.MessageTypeID,
				tg.PageBlockCoverTypeID,
				tg.InputMediaUploadedDocumentTypeID:
				allocThreshold = 256
			}
		}

		testutil.MaxAlloc(t, allocThreshold, func() {
			b.ResetTo(data)
			_ = c.handleMessage(0, b)
		})
	})
}
