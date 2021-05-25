package mtproto

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/neo"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/internal/testutil"
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

type dohuyaData struct {
	Data []byte
}

func (d dohuyaData) Decode(b *bin.Buffer) error {
	_, err := b.Bytes()
	return err
}

func (d dohuyaData) Encode(b *bin.Buffer) error {
	b.PutBytes(d.Data)
	return nil
}

type noopBuf struct{}

func (n noopBuf) Consume(id int64) bool {
	return true
}

type constantConn struct {
	data []byte
}

func (c constantConn) Send(ctx context.Context, b *bin.Buffer) error {
	return nil
}

func (c constantConn) Recv(ctx context.Context, b *bin.Buffer) error {
	b.Put(c.data)
	return nil
}

func (c constantConn) Close() error {
	return nil
}

func benchRead(payloadSize int) func(b *testing.B) {
	return func(b *testing.B) {
		a := require.New(b)
		procs := runtime.GOMAXPROCS(0)
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
		a.NoError(msg.Encode(&dohuyaData{
			Data: payload,
		}))

		length := msg.Len()
		data := msg.Copy()
		a.NoError(serverCipher.Encrypt(authKey, crypto.EncryptedMessageData{
			MessageID:              id,
			SeqNo:                  1,
			MessageDataLen:         int32(length),
			MessageDataWithPadding: data,
		}, msg))

		ackCh := make(chan int64, procs)
		conn := Conn{
			conn:            constantConn{data: msg.Raw()},
			handler:         nopHandler{},
			clock:           c,
			rand:            random,
			cipher:          crypto.NewClientCipher(random),
			log:             logger,
			messageID:       proto.NewMessageIDGen(c.Now),
			messageIDBuf:    noopBuf{},
			ackSendChan:     ackCh,
			authKey:         authKey,
			readConcurrency: procs,
			messages:        make(chan *crypto.EncryptedMessageData, procs),
		}
		grp := tdsync.NewCancellableGroup(context.Background())

		b.ResetTimer()
		b.ReportAllocs()

		grp.Go(conn.readLoop)
		for i := 0; i < procs; i++ {
			grp.Go(conn.readEncryptedMessages)
		}

		grp.Go(func(ctx context.Context) error {
			defer grp.Cancel()

			for i := 0; i < b.N; i++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-ackCh:
				}
			}

			return nil
		})
		a.ErrorIs(grp.Wait(), context.Canceled)
	}
}

func BenchmarkRead(b *testing.B) {
	for _, size := range []int{
		128,
		1024,
		16 * 1024,
		512 * 1024,
	} {
		b.Run(fmt.Sprintf("%db", size), benchRead(size))
	}
}
