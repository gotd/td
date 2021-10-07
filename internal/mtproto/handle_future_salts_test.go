package mtproto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/mt"
)

func TestConn_handleFutureSalts(t *testing.T) {
	now := time.Now()
	ts := int(now.Unix())
	testdata := []mt.FutureSalt{
		{
			ValidSince: ts - 1,
			ValidUntil: ts + 1,
			Salt:       10,
		},
		{
			ValidSince: ts + 1,
			ValidUntil: ts + 3,
			Salt:       11,
		},
	}

	t.Run("OK", func(t *testing.T) {
		a := require.New(t)
		conn := Conn{log: zap.NewNop()}
		buf := bin.Buffer{}

		a.NoError(buf.Encode(&mt.FutureSalts{
			ReqMsgID: 1,
			Now:      1,
			Salts:    testdata,
		}))
		a.NoError(conn.handleFutureSalts(&buf))

		salt, ok := conn.salts.Get(now)
		a.Equal(int64(10), salt)
		a.True(ok)
	})
	t.Run("Invalid", func(t *testing.T) {
		conn := Conn{}
		buf := bin.Buffer{}
		buf.PutID(mt.FutureSaltsTypeID)
		require.Error(t, conn.handleFutureSalts(&buf))
	})
}
