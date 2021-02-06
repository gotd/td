package pool

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/internal/tdsync"
)

func TestReqMap(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grp := tdsync.NewCancellableGroup(ctx)
	req := newReqMap()

	key, ch := req.request()
	grp.Go(func(groupCtx context.Context) error {
		defer req.delete(key)

		select {
		case <-ch:
			return nil
		case <-groupCtx.Done():
			return groupCtx.Err()
		}
	})

	grp.Go(func(groupCtx context.Context) error {
		require.True(t, req.transfer(&poolConn{}))
		require.False(t, req.transfer(&poolConn{}))
		return nil
	})

	require.NoError(t, grp.Wait())
}
