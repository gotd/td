package inline

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/internal/rpcmock"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func testBuilder(t *testing.T) (*ResultBuilder, *rpcmock.Mock) {
	mock := rpcmock.NewMock(t, require.New(t))
	sender := New(tg.NewClient(mock), rand.Reader, 10)
	return sender, mock
}

func testRPCError() *tgerr.Error {
	return &tgerr.Error{
		Code:    1337,
		Message: "TEST_ERROR",
		Type:    "TEST_ERROR",
	}
}
