package pool

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/internal/tdsync"
)

func TestReqMap(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	g := tdsync.NewCancellableGroup(ctx)
	req := newReqMap()

	key, ch := req.request()
	g.Go(func(ctx context.Context) error {
		defer req.delete(key)

		select {
		case <-ch:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	g.Go(func(ctx context.Context) error {
		require.True(t, req.transfer(&poolConn{}))
		require.False(t, req.transfer(&poolConn{}))
		return nil
	})

	require.NoError(t, g.Wait())
}
