package main

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.uber.org/zap"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/examples"
	"github.com/nnqq/td/tdp"
	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/tg"
)

// prettyMiddleware pretty-prints request and response.
func prettyMiddleware() telegram.MiddlewareFunc {
	return func(next tg.Invoker) telegram.InvokeFunc {
		return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
			fmt.Println("→", formatObject(input))
			start := time.Now()
			if err := next.Invoke(ctx, input, output); err != nil {
				fmt.Println("←", err)
				return err
			}

			fmt.Printf("← (%s) %s\n", time.Since(start).Round(time.Millisecond), formatObject(output))

			return nil
		}
	}
}

func formatObject(input interface{}) string {
	o, ok := input.(tdp.Object)
	if !ok {
		// Handle tg.*Box values.
		rv := reflect.Indirect(reflect.ValueOf(input))
		for i := 0; i < rv.NumField(); i++ {
			if v, ok := rv.Field(i).Interface().(tdp.Object); ok {
				return formatObject(v)
			}
		}

		return fmt.Sprintf("%T (not object)", input)
	}
	return tdp.Format(o)
}

func main() {
	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		return telegram.BotFromEnvironment(ctx, telegram.Options{
			UpdateHandler: telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
				// Print all incoming updates.
				fmt.Println("u:", formatObject(u))
				return nil
			}),
			Middlewares: []telegram.Middleware{
				prettyMiddleware(),
			},
		}, nil, telegram.RunUntilCanceled)
	})
}
