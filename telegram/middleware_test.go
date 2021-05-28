package telegram

import (
	"context"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

type testPrintInvoker struct {
	t *testing.T
}

func (t testPrintInvoker) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	t.t.Log("invoke")
	return nil
}

func TestExampleMiddleware(t *testing.T) {
	_ = chainMiddlewares(testPrintInvoker{t: t},
		MiddlewareFunc(func(next tg.Invoker) InvokeFunc {
			return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
				t.Log("First")
				defer t.Log("After first")
				return next.Invoke(ctx, input, output)
			}
		}),
		MiddlewareFunc(func(next tg.Invoker) InvokeFunc {
			return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
				t.Log("Second")
				defer t.Log("After second")
				return next.Invoke(ctx, input, output)
			}
		}),
	).Invoke(context.Background(), nil, nil)
}
