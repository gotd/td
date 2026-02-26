package telegram

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
)

func TestClientHandleDCConnDeadPFSDropResetsSession(t *testing.T) {
	a := require.New(t)
	client := Client{
		log: zap.NewNop(),
	}
	client.init()

	dcID := 5
	key := crypto.Key{1}.WithID()
	session := pool.NewSyncSession(pool.Session{
		DC:      dcID,
		AuthKey: key,
		Salt:    77,
	})
	client.sessions[dcID] = session

	onDeadCalls := 0
	client.onDead = func(error) {
		onDeadCalls++
	}

	client.handleDCConnDead(dcID, mtproto.ErrPFSDropKeysRequired)

	// PFS drop request forces cached key reset.
	data := session.Load()
	a.True(data.AuthKey.Zero())
	a.Zero(data.Salt)
	a.Equal(1, onDeadCalls)
}

func TestClientHandleDCConnDeadPassThrough(t *testing.T) {
	a := require.New(t)
	client := Client{
		log: zap.NewNop(),
	}
	client.init()

	dcID := 6
	key := crypto.Key{2}.WithID()
	session := pool.NewSyncSession(pool.Session{
		DC:      dcID,
		AuthKey: key,
		Salt:    88,
	})
	client.sessions[dcID] = session

	testErr := errors.New("test")
	onDeadCalls := 0
	client.onDead = func(err error) {
		a.Equal(testErr, err)
		onDeadCalls++
	}

	client.handleDCConnDead(dcID, testErr)

	// Non-PFS error must not mutate stored auth key.
	data := session.Load()
	a.Equal(key, data.AuthKey)
	a.Equal(int64(88), data.Salt)
	a.Equal(1, onDeadCalls)
}
