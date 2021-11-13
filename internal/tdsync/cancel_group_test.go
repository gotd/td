package tdsync

import (
	"context"
	"testing"

	"github.com/go-faster/errors"
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
