package telegram

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
	"github.com/gotd/td/tg"
)

func TestClientOnCDNSessionStoresSeparateMap(t *testing.T) {
	a := require.New(t)
	c := &Client{
		log: zap.NewNop(),
	}
	c.init()

	const dcID = 7
	regularKey := crypto.Key{1}.WithID()
	cdnKey := crypto.Key{2}.WithID()

	c.sessions[dcID] = pool.NewSyncSession(pool.Session{
		DC:      dcID,
		AuthKey: regularKey,
		Salt:    11,
	})

	err := c.onCDNSession(tg.Config{ThisDC: dcID}, mtproto.Session{
		Key:  cdnKey,
		Salt: 22,
	})
	a.NoError(err)

	regular := c.sessions[dcID].Load()
	a.Equal(regularKey, regular.AuthKey)
	a.Equal(int64(11), regular.Salt)

	cdn, ok := c.cdnSessions[dcID]
	a.True(ok)
	cdnData := cdn.Load()
	a.Equal(cdnKey, cdnData.AuthKey)
	a.Equal(int64(22), cdnData.Salt)
}

func TestCDNHandlerUsesCDNSessionPath(t *testing.T) {
	a := require.New(t)
	c := &Client{
		log: zap.NewNop(),
	}
	c.init()

	const dcID = 8
	cdnKey := crypto.Key{3}.WithID()

	h := c.asCDNHandler()
	err := h.OnSession(tg.Config{ThisDC: dcID}, mtproto.Session{
		Key:  cdnKey,
		Salt: 33,
	})
	a.NoError(err)

	_, regularOk := c.sessions[dcID]
	a.False(regularOk)

	cdn, cdnOK := c.cdnSessions[dcID]
	a.True(cdnOK)
	a.Equal(cdnKey, cdn.Load().AuthKey)
	a.Equal(int64(33), cdn.Load().Salt)
}
