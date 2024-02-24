//go:build go1.18

package exchange

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/testutil"
	"github.com/gotd/td/transport"
)

func FuzzValid(f *testing.F) {
	f.Add([]byte{1, 2, 3})
	f.Fuzz(func(t *testing.T, data []byte) {
		const dc = 2
		reader := testutil.Rand(data)
		privateKey := PrivateKey{
			RSA: testutil.RSAPrivateKey(),
		}

		config := zap.NewProductionConfig()
		config.OutputPaths = []string{"stdout"}
		log, err := config.Build()
		if err != nil {
			t.Fatal(err)
		}

		i := transport.Intermediate
		client, server := i.Pipe()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		g, gctx := errgroup.WithContext(ctx)
		g.Go(func() error {
			_, err := NewExchanger(client, dc).
				WithLogger(log.Named("client")).
				WithRand(reader).
				Client([]PublicKey{privateKey.Public()}).
				Run(gctx)
			if err != nil {
				cancel()
			}
			return err
		})
		g.Go(func() error {
			_, err := NewExchanger(server, dc).
				WithLogger(log.Named("server")).
				WithRand(reader).
				Server(privateKey).
				Run(gctx)
			if err != nil {
				cancel()
			}
			return err
		})

		if err := g.Wait(); err != nil {
			t.Fatal(err)
		}
	})
}
