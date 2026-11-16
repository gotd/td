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

// trackedParent tracks child registrations via the context.AfterFunc protocol.
type trackedParent struct {
	context.Context
	done     chan struct{}
	children int
}

func (p *trackedParent) Done() <-chan struct{} { return p.done }

func (p *trackedParent) AfterFunc(func()) func() bool {
	p.children++
	stopped := false
	return func() bool {
		if stopped {
			return false
		}
		stopped = true
		p.children--
		return true
	}
}

func TestCancellableGroupWaitReleasesParent(t *testing.T) {
	failure := errors.New("connection failed")
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "success"},
		{name: "failure", err: failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parent := &trackedParent{Context: context.Background(), done: make(chan struct{})}
			for i := 0; i < 100; i++ {
				g := NewCancellableGroup(parent)
				if parent.children != 1 {
					t.Fatalf("expected one registered child, got %d", parent.children)
				}
				g.Go(func(context.Context) error { return tc.err })
				if err := g.Wait(); !errors.Is(err, tc.err) {
					t.Fatalf("Wait() = %v, want %v", err, tc.err)
				}
				if parent.children != 0 {
					t.Fatalf("iteration %d: %d children retained after Wait", i, parent.children)
				}
				if parent.Err() != nil {
					t.Fatal("Wait canceled the parent")
				}
			}
		})
	}
}
