// +build go1.17

package exchange

import (
	"context"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/internal/testutil"
	"github.com/gotd/td/transport"
)

func FuzzFlow(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		reader := testutil.Rand(data)
		k := testutil.RSAPrivateKey()

		config := zap.NewProductionConfig()
		config.OutputPaths = []string{"stdout"}
		log, err := config.Build()
		require.NoError(t, err)

		i := transport.Intermediate
		client, server := i.Pipe()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		g, gctx := errgroup.WithContext(ctx)
		g.Go(func() error {
			_, err := NewExchanger(client).
				WithLogger(log.Named("client")).
				WithRand(reader).
				Client([]*rsa.PublicKey{&k.PublicKey}).
				Run(gctx)
			if err != nil {
				cancel()
			}
			return err
		})
		g.Go(func() error {
			_, err := NewExchanger(server).
				WithLogger(log.Named("server")).
				WithRand(reader).
				Server(k).
				Run(gctx)
			if err != nil {
				cancel()
			}
			return err
		})

		require.NoError(t, g.Wait())
	})
}
