package tdsync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// A finished group must not stay registered as a child of a long-lived parent.
func TestCancellableGroupWaitReleasesContext(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()

	g := NewCancellableGroup(parent)
	g.Go(func(ctx context.Context) error { return nil })
	require.NoError(t, g.Wait())

	select {
	case <-g.ctx.Done():
	default:
		t.Fatal("group context still live after Wait: it leaks into the parent's children")
	}
	require.NoError(t, parent.Err(), "parent must stay usable")
}
