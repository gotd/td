package mtproto

import (
	"context"
	"crypto/rand"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/neo"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/internal/proto"
	"github.com/nnqq/td/internal/tdsync"
	"github.com/nnqq/td/internal/testutil"
)

func TestCheckMessageID(t *testing.T) {
	now := testutil.Date()
	t.Run("Good", func(t *testing.T) {
		for _, good := range []proto.MessageID{
			proto.NewMessageID(now, proto.MessageFromServer),
			proto.NewMessageID(now, proto.MessageServerResponse),
			proto.NewMessageID(now.Add(time.Second*29), proto.MessageFromServer),
			proto.NewMessageID(now.Add(-time.Second*299), proto.MessageFromServer),
		} {
			t.Run(good.String(), func(t *testing.T) {
				require.NoError(t, checkMessageID(now, int64(good)))
			})
		}
	})
	t.Run("Bad", func(t *testing.T) {
		for _, bad := range []proto.MessageID{
			proto.NewMessageID(now, proto.MessageFromClient),
			proto.NewMessageID(now.Add(time.Second*31), proto.MessageFromServer),
			proto.NewMessageID(now.Add(-time.Second*301), proto.MessageFromServer),
			proto.NewMessageID(time.Time{}, proto.MessageFromServer),
			proto.NewMessageID(now.AddDate(-10, 0, 0), proto.MessageServerResponse),
			proto.NewMessageID(time.Time{}, proto.MessageFromClient),
		} {
			t.Run(bad.String(), func(t *testing.T) {
				require.ErrorIs(t, checkMessageID(now, int64(bad)), errRejected)
			})
		}
	})
}

func benchRead(payloadSize int) func(b *testing.B) {
	return func(b *testing.B) {
		a := require.New(b)
		logger := zap.NewNop()
		random := rand.Reader
		c := neo.NewTime(time.Now())

		var key crypto.Key
		_, err := io.ReadFull(random, key[:])
		a.NoError(err)
		authKey := key.WithID()

		payload := make([]byte, payloadSize)
		_, err = io.ReadFull(random, payload)
		a.NoError(err)

		msg := new(bin.Buffer)
		serverCipher := crypto.NewServerCipher(random)
		id := proto.NewMessageIDGen(c.Now).New(proto.MessageServerResponse)
		a.NoError(msg.Encode(&testPayload{
			Data: payload,
		}))

		length := msg.Len()
		data := msg.Copy()
		a.NoError(serverCipher.Encrypt(authKey, crypto.EncryptedMessageData{
			MessageID:              id,
			SeqNo:                  0,
			MessageDataLen:         int32(length),
			MessageDataWithPadding: data,
		}, msg))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		conn := Conn{
			conn: &constantConn{
				data:    msg.Raw(),
				cancel:  cancel,
				counter: b.N,
			},
			handler:           nopHandler{},
			clock:             c,
			rand:              random,
			cipher:            crypto.NewClientCipher(random),
			log:               logger,
			messageIDBuf:      noopBuf{},
			authKey:           authKey,
			compressThreshold: -1,
		}
		grp := tdsync.NewCancellableGroup(ctx)

		b.ResetTimer()
		b.ReportAllocs()
		b.SetBytes(int64(payloadSize))

		grp.Go(conn.readLoop)
		a.ErrorIs(grp.Wait(), context.Canceled)
	}
}

func BenchmarkRead(b *testing.B) {
	testutil.RunPayloads(b, benchRead)
}
