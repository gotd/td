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

func NewSuite(TB testing.TB, ctx context.Context, log *zap.Logger) Suite {
	return Suite{TB: TB, Ctx: ctx, Log: log}
}
