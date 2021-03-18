package exchange

import (
	"context"
	"crypto/rsa"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/td/internal/tdsync"
	"github.com/gotd/td/transport"
)

func TestExchangeTimeout(t *testing.T) {
	a := require.New(t)

	reader := rand.New(rand.NewSource(1))
	key, err := rsa.GenerateKey(reader, 2048)
	a.NoError(err)
	log := zaptest.NewLogger(t)

	i := transport.Intermediate()
	client, _ := i.Pipe()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	g := tdsync.NewCancellableGroup(ctx)
	g.Go(func(ctx context.Context) error {
		_, err := NewExchanger(client).
			WithLogger(log.Named("client")).
			WithRand(reader).
			WithTimeout(1 * time.Second).
			Client([]*rsa.PublicKey{&key.PublicKey}).
			Run(ctx)
		return err
	})

	err = g.Wait()
	if err, ok := err.(net.Error); !ok || !err.Timeout() {
		require.NoError(t, err)
	}
}
