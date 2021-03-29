package dialogs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/internal/rpcmock"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func TestElem(t *testing.T) {
	ctx := context.Background()
	mock := rpcmock.NewMock(t, require.New(t))
	raw := tg.NewClient(mock)

	ch := Elem{
		Peer: &tg.InputPeerChannel{},
	}
	testErr := tgerr.New(1337, "TEST_ERROR")

	var err error
	mock.Expect().ThenRPCErr(testErr)
	_, err = ch.Messages(raw).Count(ctx)
	mock.Error(err)
	mock.Expect().ThenRPCErr(testErr)
	_, err = ch.Search(raw).Count(ctx)
	mock.Error(err)
	mock.Expect().ThenRPCErr(testErr)
	_, err = ch.Replies(raw).Count(ctx)
	mock.Error(err)
	mock.Expect().ThenRPCErr(testErr)
	_, err = ch.UnreadMentions(raw).Count(ctx)
	mock.Error(err)
	mock.Expect().ThenRPCErr(testErr)
	_, err = ch.RecentLocations(raw).Count(ctx)
	mock.Error(err)

	_, ok := ch.Participants(raw)
	mock.True(ok)
	_, ok = ch.UserPhotos(raw)
	mock.False(ok)

	ch = Elem{
		Peer: &tg.InputPeerUser{},
	}

	_, ok = ch.Participants(raw)
	mock.False(ok)
	_, ok = ch.UserPhotos(raw)
	mock.True(ok)
}
