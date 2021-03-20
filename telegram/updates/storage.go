package updates

import (
	"context"
)

// Box is a abstraction for one storage entry.
type Box interface {
	Commit(ctx context.Context, pts int) error
	Load(ctx context.Context) (int, error)
}

// Storage is a abstraction for persistent storage.
type Storage interface {
	Acquire(ctx context.Context, name string, cb func(Box) error) error
	Keys(ctx context.Context) ([]string, error)
}
