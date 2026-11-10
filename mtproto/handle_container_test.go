package mtproto

import (
	"context"
	"crypto/rand"
	"io"
	"testing"

	"github.com/gotd/log"
	"github.com/gotd/neo"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/testutil"
)

func TestConnContainerAck(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	clock := neo.NewTime(testutil.Date())

	var key crypto.Key
	_, err := io.ReadFull(rand.Reader, key[:])
	a.NoError(err)
	authKey := key.WithID()

	gen := proto.NewMessageIDGen(clock.Now)
	var (
		contentID = gen.New(proto.MessageFromServer) // e.g. updates, odd seqno
		serviceID = gen.New(proto.MessageFromServer) // not content-related, even seqno
		outerID   = gen.New(proto.MessageFromServer) // container, even seqno
	)
	body := func() []byte {
		b := new(bin.Buffer)
		b.PutID(0xdeadbeef) // handled by Handler.OnMessage
		b.PutInt32(42)
		return b.Copy()
	}
	container := proto.MessageContainer{Messages: []proto.Message{
		{ID: contentID, SeqNo: 1, Bytes: len(body()), Body: body()},
		{ID: serviceID, SeqNo: 2, Bytes: len(body()), Body: body()},
	}}
	plain := new(bin.Buffer)
	a.NoError(container.Encode(plain))

	msg := new(bin.Buffer)
	a.NoError(crypto.NewServerCipher(rand.Reader).Encrypt(authKey, crypto.EncryptedMessageData{
		MessageID:              outerID,
		SeqNo:                  4,
		MessageDataLen:         int32(plain.Len()),
		MessageDataWithPadding: plain.Copy(),
	}, msg))

	acks := make(chan int64, 10)
	c := &Conn{
		handler:      nopHandler{},
		clock:        clock,
		rand:         rand.Reader,
		cipher:       crypto.NewClientCipher(rand.Reader),
		log:          log.For(log.Nop),
		messageIDBuf: proto.NewMessageIDBuf(10),
		authKey:      authKey,
		ackSendChan:  acks,
	}
	a.NoError(c.consumeMessage(ctx, msg))

	close(acks)
	var got []int64
	for id := range acks {
		got = append(got, id)
	}
	a.Equal([]int64{contentID}, got, "only the odd-seqno inner message must be acked")
}
