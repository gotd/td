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

	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/internal/tdsync"
	"github.com/nnqq/td/internal/testutil"
	"github.com/nnqq/td/transport"
)

func testExchange(rsaPad bool) func(t *testing.T) {
	return func(t *testing.T) {
		a := require.New(t)
		log := zaptest.NewLogger(t)

		dc := 2
		reader := rand.New(rand.NewSource(1))
		key, err := rsa.GenerateKey(reader, crypto.RSAKeyBits)
		a.NoError(err)
		privateKey := PrivateKey{
			RSA:       key,
			UseRSAPad: rsaPad,
		}

		i := transport.Intermediate
		client, server := i.Pipe()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		g := tdsync.NewCancellableGroup(ctx)
		g.Go(func(ctx context.Context) error {
			_, err := NewExchanger(client, dc).
				WithLogger(log.Named("client")).
				WithRand(reader).
				Client([]PublicKey{privateKey.Public()}).
				Run(ctx)
			return err
		})

		g.Go(func(ctx context.Context) error {
			_, err := NewExchanger(server, dc).
				WithLogger(log.Named("server")).
				WithRand(reader).
				Server(privateKey).
				Run(ctx)
			return err
		})

		a.NoError(g.Wait())
	}
}

func TestExchange(t *testing.T) {
	t.Run("PQInnerData", testExchange(false))
	t.Run("PQInnerDataDC", testExchange(true))
}

func TestExchangeCorpus(t *testing.T) {
	privateKey := PrivateKey{
		RSA:       testutil.RSAPrivateKey(),
		UseRSAPad: false,
	}

	for i, seed := range []string{
		"\xef\x00\x04",
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			dc := 2
			reader := testutil.Rand([]byte(seed))
			log := zaptest.NewLogger(t)

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

			require.NoError(t, g.Wait())
		})
	}
}
