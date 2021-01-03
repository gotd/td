package exchange

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/transport"
)

func TestExchange(t *testing.T) {
	a := require.New(t)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	a.NoError(err)
	log := zaptest.NewLogger(t)

	i := transport.Intermediate(nil)
	client, server := i.Pipe()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		_, err := NewExchanger(client).
			WithLogger(log.Named("client")).
			Client([]*rsa.PublicKey{&key.PublicKey}).
			Run(gctx)
		if err != nil {
			cancel()
		}
		return err
	})
	g.Go(func() error {
		_, err := NewExchanger(server).
			WithLogger(log.Named("server")).
			Server(key).
			Run(gctx)
		if err != nil {
			cancel()
		}
		return err
	})

	require.NoError(t, g.Wait())
}
