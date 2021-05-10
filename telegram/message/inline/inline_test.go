package inline

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
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

func TestResultBuilder_Set(t *testing.T) {
	ctx := context.Background()
	builder, mock := testBuilder(t)

	mock.ExpectFunc(func(b bin.Encoder) {
		v, ok := b.(*tg.MessagesSetInlineBotResultsRequest)
		mock.True(ok)
		mock.True(v.Gallery)
	}).ThenTrue()
	_, err := builder.Gallery(true).Set(ctx)
	mock.NoError(err)

	mock.ExpectFunc(func(b bin.Encoder) {
		v, ok := b.(*tg.MessagesSetInlineBotResultsRequest)
		mock.True(ok)
		mock.True(v.Private)
	}).ThenTrue()
	_, err = builder.Private(true).Set(ctx)
	mock.NoError(err)

	mock.ExpectFunc(func(b bin.Encoder) {
		v, ok := b.(*tg.MessagesSetInlineBotResultsRequest)
		mock.True(ok)
		mock.Equal(1, v.CacheTime)
	}).ThenTrue()
	_, err = builder.CacheTime(time.Second).Set(ctx)
	mock.NoError(err)

	mock.ExpectFunc(func(b bin.Encoder) {
		v, ok := b.(*tg.MessagesSetInlineBotResultsRequest)
		mock.True(ok)
		mock.Equal("offset", v.NextOffset)
	}).ThenTrue()
	_, err = builder.NextOffset("offset").Set(ctx)
	mock.NoError(err)
}
