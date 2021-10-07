package mtproto

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/neo"
	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/mt"
	"github.com/nnqq/td/internal/rpc"
	"github.com/nnqq/td/internal/tdsync"
	"github.com/nnqq/td/internal/testutil"
	"github.com/nnqq/td/internal/tmap"
	"github.com/nnqq/td/tg"
)

type testUpdateHandler struct {
	types *tmap.Constructor
}

func (h testUpdateHandler) OnMessage(b *bin.Buffer) error {
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

func (testUpdateHandler) OnSession(session Session) error { return nil }

func newTestHandler() Handler {
	return &testUpdateHandler{
		types: tmap.NewConstructor(
			tg.TypesConstructorMap(),
			mt.TypesConstructorMap(),
		),
	}
}

func TestConnHandleMessage(t *testing.T) {
	c := &Conn{
		rand:    Zero{},
		log:     zap.NewNop(),
		handler: newTestHandler(),
	}

	for i, input := range []string{
		"\xdc\xf8\xf1stewa\x00O\x03expired c" +
			"ertificate\x02\x00\x00\x00\xef",

		"\x01m\\\xf300000000\x19\xcaD!0000" +
			"\x1100000000000000000",

		"\x01m\\\xf300000000\x19\xcaD!0000" +
			"\xfe0",

		"@B\xaet\x15ĵ\x1c0000,\x8f\xf8B0000" +
			"00000000\x15ĵ\x1c0000\xff000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"00000000000000000000" +
			"000000000000",
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if err := c.handleMessage(0, &bin.Buffer{Buf: []byte(input)}); err == nil {
				t.Fatal("error expected")
			}
		})
	}
}

func TestConnHandleMessageCorpus(t *testing.T) {
	if testutil.Race {
		t.Skip("Skipped")
	}

	c := &Conn{
		handler:    newTestHandler(),
		rpc:        rpc.New(rpc.NopSend, rpc.Options{}),
		clock:      neo.NewTime(time.Now()),
		rand:       Zero{},
		log:        zap.NewNop(),
		gotSession: tdsync.NewReady(),
	}

	corpusDir := filepath.Join("..", "..", "_fuzz", "handle_message", "corpus")
	corpus, err := os.ReadDir(corpusDir)
	if err != nil {
		t.Fatal(err)
	}

	b := &bin.Buffer{}
	types := tmap.New(
		tg.TypesMap(),
		mt.TypesMap(),
	)

	for _, f := range corpus {
		t.Run(f.Name(), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(corpusDir, f.Name()))
			require.NoError(t, err)

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
					tg.InputMediaUploadedDocumentTypeID,
					tg.SecureRequiredTypeOneOfTypeID:
					allocThreshold = 256
				}
			}

			testutil.MaxAlloc(t, allocThreshold, func() {
				b.ResetTo(data)
				_ = c.handleMessage(0, b)
			})
		})
	}
}
