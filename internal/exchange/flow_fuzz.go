//go:build fuzz
// +build fuzz

package exchange

import (
	"context"
	"crypto/rsa"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/nnqq/td/internal/testutil"
	"github.com/nnqq/td/transport"
)

func FuzzFlow(data []byte) int {
	reader := testutil.Rand(data)
	k := testutil.RSAPrivateKey()

	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	log, err := config.Build()
	if err != nil {
		panic(err)
	}

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

	if err := g.Wait(); err != nil {
		panic(err)
	}

	return 1
}
