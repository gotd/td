package exchange

import (
	"context"
	"crypto/rsa"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/internal/testutil"
	"github.com/gotd/td/transport"
)

func TestExchange(t *testing.T) {
	a := require.New(t)

	reader := rand.New(rand.NewSource(1))
	key, err := rsa.GenerateKey(reader, 2048)
	a.NoError(err)
	log := zaptest.NewLogger(t)

	i := transport.Intermediate(nil)
	client, server := i.Pipe()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	g := tdsync.NewCancellableGroup(ctx)
	g.Go(func(ctx context.Context) error {
		_, err := NewExchanger(client).
			WithLogger(log.Named("client")).
			WithRand(reader).
			Client([]*rsa.PublicKey{&key.PublicKey}).
			Run(ctx)
		return err
	})
	g.Go(func(ctx context.Context) error {
		_, err := NewExchanger(server).
			WithLogger(log.Named("server")).
			WithRand(reader).
			Server(key).
			Run(ctx)
		return err
	})

	require.NoError(t, g.Wait())
}

func TestExchangeCorpus(t *testing.T) {
	k := testutil.RSAPrivateKey()

	for i, seed := range []string{
		"\xef\x00\x04",
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			reader := testutil.Rand([]byte(seed))
			log := zaptest.NewLogger(t)

			i := transport.Intermediate(nil)
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
}
