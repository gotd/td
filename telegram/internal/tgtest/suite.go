package tgtest

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

type Suite struct {
	testing.TB
	Ctx context.Context
	Log *zap.Logger
}

func NewSuite(ctx context.Context, tb testing.TB, log *zap.Logger) Suite {
	return Suite{TB: tb, Ctx: ctx, Log: log}
}
