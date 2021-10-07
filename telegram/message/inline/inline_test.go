package inline

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgerr"
	"github.com/nnqq/td/tgmock"
)

func testBuilder(t *testing.T) (*ResultBuilder, *tgmock.Mock) {
	mock := tgmock.New(t)
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
		require.True(t, ok)
		require.True(t, v.Gallery)
	}).ThenTrue()
	_, err := builder.Gallery(true).Set(ctx)
	require.NoError(t, err)

	mock.ExpectFunc(func(b bin.Encoder) {
		v, ok := b.(*tg.MessagesSetInlineBotResultsRequest)
		require.True(t, ok)
		require.True(t, v.Private)
	}).ThenTrue()
	_, err = builder.Private(true).Set(ctx)
	require.NoError(t, err)

	mock.ExpectFunc(func(b bin.Encoder) {
		v, ok := b.(*tg.MessagesSetInlineBotResultsRequest)
		require.True(t, ok)
		require.Equal(t, 1, v.CacheTime)
	}).ThenTrue()
	_, err = builder.CacheTime(time.Second).Set(ctx)
	require.NoError(t, err)

	mock.ExpectFunc(func(b bin.Encoder) {
		v, ok := b.(*tg.MessagesSetInlineBotResultsRequest)
		require.True(t, ok)
		require.Equal(t, "offset", v.NextOffset)
	}).ThenTrue()
	_, err = builder.NextOffset("offset").Set(ctx)
	require.NoError(t, err)
}
