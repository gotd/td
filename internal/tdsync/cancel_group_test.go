package tdsync

import (
	"context"
	"errors"
	"testing"
)

func TestCancellableGroup(t *testing.T) {
	g := NewCancellableGroup(context.Background())

	g.Go(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})

	g.Cancel()
	if err := g.Wait(); !errors.Is(err, context.Canceled) {
		t.Error(err)
	}
}
