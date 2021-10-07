package dialogs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgerr"
	"github.com/nnqq/td/tgmock"
)

func TestElem(t *testing.T) {
	ctx := context.Background()
	mock := tgmock.NewRequire(t)
	raw := tg.NewClient(mock)

	ch := Elem{
		Peer: &tg.InputPeerChannel{},
	}
	testErr := tgerr.New(1337, "TEST_ERROR")

	var err error
	mock.Expect().ThenRPCErr(testErr)
	_, err = ch.Messages(raw).Count(ctx)
	require.Error(t, err)
	mock.Expect().ThenRPCErr(testErr)
	_, err = ch.Search(raw).Count(ctx)
	require.Error(t, err)
	mock.Expect().ThenRPCErr(testErr)
	_, err = ch.Replies(raw).Count(ctx)
	require.Error(t, err)
	mock.Expect().ThenRPCErr(testErr)
	_, err = ch.UnreadMentions(raw).Count(ctx)
	require.Error(t, err)
	mock.Expect().ThenRPCErr(testErr)
	_, err = ch.RecentLocations(raw).Count(ctx)
	require.Error(t, err)

	_, ok := ch.Participants(raw)
	require.True(t, ok)
	_, ok = ch.UserPhotos(raw)
	require.False(t, ok)

	ch = Elem{
		Peer: &tg.InputPeerUser{},
	}

	_, ok = ch.Participants(raw)
	require.False(t, ok)
	_, ok = ch.UserPhotos(raw)
	require.True(t, ok)
}
