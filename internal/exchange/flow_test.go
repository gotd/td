package exchange

import (
	"context"
	"crypto/rsa"
	"fmt"
	"math/rand"
	"net"
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

	grp := tdsync.NewCancellableGroup(ctx)
	grp.Go(func(groupCtx context.Context) error {
		_, err := NewExchanger(client).
			WithLogger(log.Named("client")).
			WithRand(reader).
			Client([]*rsa.PublicKey{&key.PublicKey}).
			Run(groupCtx)
		return err
	})
	grp.Go(func(groupCtx context.Context) error {
		_, err := NewExchanger(server).
			WithLogger(log.Named("server")).
			WithRand(reader).
			Server(key).
			Run(groupCtx)
		return err
	})

	require.NoError(t, grp.Wait())
}

func TestExchangeTimeout(t *testing.T) {
	a := require.New(t)

	reader := rand.New(rand.NewSource(1))
	key, err := rsa.GenerateKey(reader, 2048)
	a.NoError(err)
	log := zaptest.NewLogger(t)

	i := transport.Intermediate(nil)
	client, _ := i.Pipe()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	grp := tdsync.NewCancellableGroup(ctx)
	grp.Go(func(groupCtx context.Context) error {
		_, err := NewExchanger(client).
			WithLogger(log.Named("client")).
			WithRand(reader).
			WithTimeout(1 * time.Second).
			Client([]*rsa.PublicKey{&key.PublicKey}).
			Run(groupCtx)
		return err
	})

	err = grp.Wait()
	if err, ok := err.(net.Error); !ok || !err.Timeout() {
		require.NoError(t, err)
	}
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
