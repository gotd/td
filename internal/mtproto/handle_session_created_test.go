package mtproto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/gotd/neo"
	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/mt"
	"github.com/nnqq/td/internal/proto"
	"github.com/nnqq/td/internal/tdsync"
)

func TestConn_handleSessionCreated(t *testing.T) {
	t.Run("NeedSynchronization", func(t *testing.T) {
		a := require.New(t)
		logger, logs := observer.New(zapcore.WarnLevel)

		now := time.Unix(100, 0)
		clock := neo.NewTime(now)
		gotSession := tdsync.NewReady()
		conn := Conn{
			clock:      clock,
			log:        zap.New(logger),
			gotSession: gotSession,
			handler:    newTestHandler(),
		}

		buf := bin.Buffer{}
		msgID := proto.NewMessageID(now.Add(maxFuture+time.Second), proto.MessageFromClient)
		a.NoError(buf.Encode(&mt.NewSessionCreated{
			FirstMsgID: int64(msgID),
			UniqueID:   10,
			ServerSalt: 10,
		}))
		a.NoError(conn.handleSessionCreated(&buf))

		select {
		case <-gotSession.Ready():
		default:
			t.Fatal("expected gotSession signal")
		}
		a.Equal(int64(10), conn.salt)

		msgs := logs.All()
		a.Len(msgs, 1)
		a.Equal("Local clock needs synchronization", msgs[0].Message)
	})
	t.Run("Invalid", func(t *testing.T) {
		conn := Conn{}
		buf := bin.Buffer{}
		buf.PutID(mt.NewSessionCreatedTypeID)
		require.Error(t, conn.handleSessionCreated(&buf))
	})
}
