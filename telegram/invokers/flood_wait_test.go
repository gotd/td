package invokers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/telegram/internal/rpcmock"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func TestWaiter(t *testing.T) {
	ctx := context.Background()
	mock := rpcmock.NewMock(t, require.New(t))
	waiter := NewWaiter(mock)
	raw := tg.NewClient(waiter)

	grp := tdsync.NewCancellableGroup(ctx)
	grp.Go(waiter.Run)
	grp.Go(func(ctx context.Context) error {
		defer grp.Cancel()

		start := time.Now()
		mock.Expect().N(2).ThenFlood(3)
		mock.Expect().ThenResult(&tg.Config{})

		_, err := raw.HelpGetConfig(ctx)
		mock.NoError(err)
		mock.GreaterOrEqualf(time.Since(start), 6*time.Second, "waiter does not wait enough")

		start = time.Now()
		mock.Expect().ThenResult(&tg.Config{})
		mock.Expect().ThenResult(&tg.Config{})
		mock.Expect().ThenResult(&tg.Config{})
		for range [3]struct{}{} {
			_, err = raw.HelpGetConfig(ctx)
			mock.NoError(err)
		}
		mock.LessOrEqualf(time.Since(start), 9*time.Second, "timer does not decrease")

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

		return nil
	})

	_ = grp.Wait()
}
