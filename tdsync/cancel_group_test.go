package tdsync

import (
	"context"
	"errors"
	"sync/atomic"
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

// cancelTrackingContext uses the public AfterFunc hook to observe whether a
// child has detached from its parent, without inspecting context internals.
type cancelTrackingContext struct {
	context.Context
	children atomic.Int32
}

func (c *cancelTrackingContext) Value(key any) any {
	// Hide the underlying cancelCtx so context uses AfterFunc to subscribe.
	return nil
}

func (c *cancelTrackingContext) AfterFunc(f func()) func() bool {
	c.children.Add(1)
	stop := context.AfterFunc(c.Context, f)
	return func() bool {
		if !stop() {
			return false
		}
		c.children.Add(-1)
		return true
	}
}

func TestCancellableGroupWaitReleasesParent(t *testing.T) {
	workerErr := errors.New("worker failed")
	for _, tc := range []struct {
		name   string
		empty  bool
		cancel bool
		err    error
	}{
		{name: "empty", empty: true},
		{name: "success"},
		{name: "error", err: workerErr},
		{name: "explicit cancel", cancel: true, err: context.Canceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			parent := &cancelTrackingContext{Context: ctx}
			g := NewCancellableGroup(parent)
			defer g.Cancel()

			if got := parent.children.Load(); got != 1 {
				t.Fatalf("active parent subscriptions = %d, want 1", got)
			}
			if !tc.empty {
				g.Go(func(ctx context.Context) error {
					if tc.cancel {
						<-ctx.Done()
						return ctx.Err()
					}
					return tc.err
				})
			}
			if tc.cancel {
				g.Cancel()
			}

			if err := g.Wait(); !errors.Is(err, tc.err) {
				t.Fatalf("Wait() = %v, want %v", err, tc.err)
			}
			if got := parent.children.Load(); got != 0 {
				t.Errorf("parent subscriptions after Wait() = %d, want 0", got)
			}
			if err := parent.Err(); err != nil {
				t.Errorf("Wait canceled the parent: %v", err)
			}
			if err := g.ctx.Err(); !errors.Is(err, context.Canceled) {
				t.Errorf("group context error = %v, want context.Canceled", err)
			}
			if err := g.Wait(); !errors.Is(err, tc.err) {
				t.Errorf("second Wait() = %v, want %v", err, tc.err)
			}
		})
	}
}
