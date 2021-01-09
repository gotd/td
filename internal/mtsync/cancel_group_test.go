package mtsync

import (
	"context"
	"errors"
	"testing"
)

func TestCancellableGroup(t *testing.T) {
	grp := NewCancellableGroup(context.Background())

	grp.Go(func(groupCtx context.Context) error {
		<-groupCtx.Done()
		return groupCtx.Err()
	})

	grp.Cancel()
	if err := grp.Wait(); !errors.Is(err, context.Canceled) {
		t.Error(err)
	}
}
