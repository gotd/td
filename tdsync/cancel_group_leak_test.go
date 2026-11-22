package tdsync

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

// countingParent is a long-lived parent context that counts the children
// registered on it. context.WithCancel registers a child through AfterFunc
// when the parent provides it, and calls the returned stop func when the
// child is cancelled, so live is the number of children not yet released.
type countingParent struct {
	context.Context
	live atomic.Int64
}

// Value hides the embedded cancelCtx, so context.WithCancel registers children
// through AfterFunc instead of directly in the embedded context.
func (p *countingParent) Value(any) any { return nil }

func (p *countingParent) AfterFunc(f func()) func() bool {
	p.live.Add(1)
	stop := context.AfterFunc(p.Context, f)
	var once sync.Once
	return func() bool {
		once.Do(func() { p.live.Add(-1) })
		return stop()
	}
}

// A finished group must not stay registered as a child of a long-lived
// parent: before Wait cancelled the group context, every group started from
// the client's reconnect loop, and every download and upload, left one
// context in its parent until the parent itself was cancelled.
func TestCancellableGroupWaitReleasesContext(t *testing.T) {
	base, cancel := context.WithCancel(context.Background())
	defer cancel()
	parent := &countingParent{Context: base}

	for i := 0; i < 100; i++ {
		g := NewCancellableGroup(parent)
		g.Go(func(ctx context.Context) error { return nil })
		require.NoError(t, g.Wait())
	}

	require.Zero(t, parent.live.Load(), "finished groups still registered in their parent")
	require.NoError(t, parent.Err(), "parent must stay usable")
}
