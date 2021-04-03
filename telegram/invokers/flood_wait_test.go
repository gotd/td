package invokers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/internal/rpcmock"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func TestWaiter(t *testing.T) {
	ctx := context.Background()
	mock := rpcmock.NewMock(t, require.New(t))
	waiter := NewWaiter(mock)
	raw := tg.NewClient(waiter)

	mock.Expect().N(2).ThenFlood(3)
	mock.Expect().ThenResult(&tg.Config{})
	mock.Expect().ThenResult(&tg.Config{})
	mock.Expect().ThenResult(&tg.Config{})
	mock.Expect().ThenResult(&tg.Config{})
	mock.Expect().N(1).ThenFlood(3)
	mock.Expect().ThenResult(&tg.Config{})
	mock.Expect().ThenResult(&tg.Config{})
	mock.Expect().ThenResult(&tg.Config{})
	mock.Expect().N(3).ThenFlood(2)
	mock.Expect().ThenResult(&tg.Config{})
	mock.Expect().ThenRPCErr(tgerr.New(1337, "TEST_ERROR"))

	for {
		_, err := raw.HelpGetConfig(ctx)
		if tgerr.IsCode(err, 1337) {
			break
		}
		mock.NoError(err)
	}
}
