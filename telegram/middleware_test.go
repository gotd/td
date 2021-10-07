package telegram

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/tg"
)

func TestMiddlewareOrder(t *testing.T) {
	var calls []int
	call := func(i int) {
		calls = append(calls, i)
	}
	_ = chainMiddlewares(
		// Client.
		InvokeFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
			call(0)
			return nil
		}),
		// First middleware (index = 0).
		MiddlewareFunc(func(next tg.Invoker) InvokeFunc {
			return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
				t.Log("First")
				call(1)
				defer call(-1)
				return next.Invoke(ctx, input, output)
			}
		}),
		// Second middleware (index = 1).
		MiddlewareFunc(func(next tg.Invoker) InvokeFunc {
			return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
				call(2)
				defer call(-2)
				return next.Invoke(ctx, input, output)
			}
		}),
	).Invoke(context.Background(), nil, nil)
	require.Equal(t, []int{1, 2, 0, -2, -1}, calls)
}

func ExampleMiddleware() {
	invoker := InvokeFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		fmt.Println("invoke")
		return nil
	})
	printMiddleware := func(message string) Middleware {
		return MiddlewareFunc(func(next tg.Invoker) InvokeFunc {
			return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
				fmt.Println(message, "(start)")
				err := next.Invoke(ctx, input, output)
				fmt.Println(message, "(end)")
				return err
			}
		})
	}

	// Testing composed invoker.
	_ = chainMiddlewares(invoker,
		printMiddleware("first"),
		printMiddleware("second"),
	).Invoke(context.Background(), nil, nil)

	// Output:
	// first (start)
	// second (start)
	// invoke
	// second (end)
	// first (end)
}
