package examples

import (
	"context"

	"go.uber.org/zap"
)

// Run runs f callback with context and logger, panics on error.
func Run(f func(ctx context.Context, log *zap.Logger) error) {
	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()
	// No graceful shutdown.
	ctx := context.Background()
	if err := f(ctx, log); err != nil {
		log.Fatal("Run failed", zap.Error(err))
	}
	// Done.
}
