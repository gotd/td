package mtproto

import (
	"context"
	"crypto/rand"
	"io"
	"runtime"
	"testing"
	"time"

	"github.com/gotd/neo"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/internal/testutil"
	"github.com/gotd/td/transport"
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

func BenchmarkRead(b *testing.B) {
	a := require.New(b)
	procs := runtime.GOMAXPROCS(0)
	logger := zap.NewNop()
	random := rand.Reader
	c := neo.NewTime(time.Now())

	payload := make([]byte, 1024)
	_, err := io.ReadFull(random, payload)
	a.NoError(err)

	var key crypto.Key
	_, err = io.ReadFull(random, key[:])
	a.NoError(err)
	authKey := key.WithID()

	ackCh := make(chan int64, procs)
	client, server := transport.PaddedIntermediate.Pipe()
	conn := Conn{
		conn:            client,
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

	ctx := context.TODO()
	grp := tdsync.NewCancellableGroup(ctx)

	msg := new(bin.Buffer)
	serverCipher := crypto.NewServerCipher(random)
	gen := proto.NewMessageIDGen(c.Now)
	id := gen.New(proto.MessageServerResponse)
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

	b.ResetTimer()
	b.ReportAllocs()

	grp.Go(conn.readLoop)
	for i := 0; i < procs; i++ {
		grp.Go(conn.readEncryptedMessages)
	}
	grp.Go(func(ctx context.Context) error {
		for i := 0; i < b.N; i++ {
			if err := server.Send(ctx, &bin.Buffer{Buf: msg.Buf}); err != nil {
				return xerrors.Errorf("send: %w", err)
			}
		}
		return nil
	})

	for i := 0; i < b.N; i++ {
		select {
		case <-ctx.Done():
		case <-ackCh:
		}
	}

	a.NoError(server.Close())
	_ = grp.Wait()
}
