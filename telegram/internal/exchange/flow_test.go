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

	randSource := rand.Reader
	clock := time.Now
	key, err := rsa.GenerateKey(randSource, 2048)
	a.NoError(err)
	log := zaptest.NewLogger(t)

	i := transport.Intermediate(nil)
	client, server := i.Pipe()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		cfg := NewConfig(clock, randSource, client, log.Named("client"))
		_, err := NewClientExchange(cfg, &key.PublicKey).Run(gctx)
		if err != nil {
			cancel()
		}
		return err
	})
	g.Go(func() error {
		cfg := NewConfig(clock, randSource, server, log.Named("server"))
		_, err := NewServerExchange(cfg, key).Run(ctx, nil)
		if err != nil {
			cancel()
		}
		return err
	})

	require.NoError(t, g.Wait())
}
