package mtproto

import (
	"bytes"
	stdgzip "compress/gzip"
	"context"
	"testing"
	"time"

	"github.com/gotd/log"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/testutil"
)

type infoHandler struct {
	nopHandler
	infos []MessageInfo
}

func (h *infoHandler) OnMessageInfo(info MessageInfo, _ *bin.Buffer) error {
	h.infos = append(h.infos, info)
	return nil
}

// The handler gets the msg_id of the inner message and the receive time of
// the outer one, through containers and gzip.
func TestConnMessageInfo(t *testing.T) {
	a := require.New(t)
	received := testutil.Date()

	body := new(bin.Buffer)
	body.PutID(0xdeadbeef) // handled by Handler
	body.PutInt32(42)

	var packed bytes.Buffer
	w := stdgzip.NewWriter(&packed)
	_, err := w.Write(body.Buf)
	a.NoError(err)
	a.NoError(w.Close())
	zipped := new(bin.Buffer)
	a.NoError(zipped.Encode(&proto.GZIP{Data: packed.Bytes()}))

	const plainID, zippedID, outerID = 100, 200, 300
	container := proto.MessageContainer{Messages: []proto.Message{
		{ID: plainID, SeqNo: 2, Bytes: body.Len(), Body: body.Copy()},
		{ID: zippedID, SeqNo: 2, Bytes: zipped.Len(), Body: zipped.Copy()},
	}}
	b := new(bin.Buffer)
	a.NoError(container.Encode(b))

	h := &infoHandler{}
	const offset = 3 * time.Second
	c := &Conn{
		handler: h, infoHandler: h, log: log.For(log.Nop),
		serverTime: NewServerTime(nil, offset),
	}
	a.NoError(c.handleMessage(context.Background(), received, outerID, b))
	a.Equal([]MessageInfo{
		{ID: plainID, Received: received, ServerTimeOffset: offset},
		{ID: zippedID, Received: received, ServerTimeOffset: offset},
	}, h.infos)

	// New picks the extension up, and falls back to OnMessage without it.
	a.NotNil(New(nil, Options{Handler: h}).infoHandler)
	a.Nil(New(nil, Options{Handler: nopHandler{}}).infoHandler)
	a.NoError((&Conn{handler: nopHandler{}, log: log.For(log.Nop)}).
		handleMessage(context.Background(), time.Time{}, plainID, body))
}
