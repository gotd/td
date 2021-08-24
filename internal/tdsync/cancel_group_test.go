package tdsync

import (
	"context"
	"testing"

	"golang.org/x/xerrors"
)

func TestCancellableGroup(t *testing.T) {
	g := NewCancellableGroup(context.Background())

	g.Go(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})

	g.Cancel()
	if err := g.Wait(); !xerrors.Is(err, context.Canceled) {
		t.Error(err)
	}
}
