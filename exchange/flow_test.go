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

	"github.com/gotd/log/logzap"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/testutil"
	"github.com/gotd/td/transport"
)

func testExchange(tempMode bool) func(t *testing.T) {
	return func(t *testing.T) {
		a := require.New(t)
		log := zaptest.NewLogger(t)

		dc := 2
		reader := rand.New(rand.NewSource(1))
		key, err := rsa.GenerateKey(reader, crypto.RSAKeyBits)
		a.NoError(err)
		privateKey := PrivateKey{
			RSA: key,
		}

		i := transport.Intermediate
		client, server := i.Pipe()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		var clientResult ClientExchangeResult
		g := tdsync.NewCancellableGroup(ctx)
		g.Go(func(ctx context.Context) error {
			ex := NewExchanger(client, dc).
				WithLogger(logzap.New(log.Named("client"))).
				WithRand(reader)
			if tempMode {
				// Force temporary key path to ensure p_q_inner_data_temp_dc wiring.
				ex = ex.WithTempMode(60)
			}
			r, err := ex.Client([]PublicKey{privateKey.Public()}).Run(ctx)
			clientResult = r
			return err
		})

		g.Go(func(ctx context.Context) error {
			_, err := NewExchanger(server, dc).
				WithLogger(logzap.New(log.Named("server"))).
				WithRand(reader).
				Server(privateKey).
				Run(ctx)
			return err
		})

		a.NoError(g.Wait())
		if tempMode {
			a.NotZero(clientResult.ExpiresAt)
		} else {
			a.Zero(clientResult.ExpiresAt)
		}
	}
}

func TestExchange(t *testing.T) {
	t.Run("PQInnerData", testExchange(false))
	t.Run("PQInnerDataDC", testExchange(true))
}

// TestServerUnexpectedEncrypted verifies that the server flow surfaces a frame
// bearing a non-zero auth key id as *UnexpectedEncryptedError instead of failing
// the exchange, so callers can resolve the key and handle it as a normal RPC
// rather than replying -404 (which makes clients discard a still-valid key).
func TestServerUnexpectedEncrypted(t *testing.T) {
	a := require.New(t)
	log := zaptest.NewLogger(t)

	dc := 2
	reader := rand.New(rand.NewSource(1))
	key, err := rsa.GenerateKey(reader, crypto.RSAKeyBits)
	a.NoError(err)
	privateKey := PrivateKey{RSA: key}

	i := transport.Intermediate
	client, server := i.Pipe()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Frame whose leading auth_key_id is non-zero: an encrypted message rather
	// than an unencrypted key-exchange message.
	authKeyID := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	var payload bin.Buffer
	payload.Put(authKeyID[:])
	// Intermediate transport requires 4-byte-aligned payloads; keep the total
	// frame length a multiple of 4.
	payload.Put([]byte("bodybyte"))
	frame := append([]byte(nil), payload.Buf...)

	var serverErr error
	g := tdsync.NewCancellableGroup(ctx)
	g.Go(func(ctx context.Context) error {
		return client.Send(ctx, &payload)
	})
	g.Go(func(ctx context.Context) error {
		_, serverErr = NewExchanger(server, dc).
			WithLogger(logzap.New(log.Named("server"))).
			WithRand(reader).
			Server(privateKey).
			Run(ctx)
		return nil
	})
	a.NoError(g.Wait())

	var encErr *UnexpectedEncryptedError
	a.ErrorAs(serverErr, &encErr)
	a.Equal(authKeyID, encErr.AuthKeyID)
	a.Equal(frame, encErr.Frame)
}

func TestExchangeCorpus(t *testing.T) {
	privateKey := PrivateKey{
		RSA: testutil.RSAPrivateKey(),
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
					WithLogger(logzap.New(log.Named("client"))).
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
					WithLogger(logzap.New(log.Named("server"))).
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
