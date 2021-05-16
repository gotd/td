package tgtest

import (
	"context"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/exchange"
	"github.com/gotd/td/transport"
)

func (s *Server) exchange(ctx context.Context, conn transport.Conn) (crypto.AuthKey, error) {
	r, err := exchange.NewExchanger(conn).
		WithClock(s.clock).
		WithLogger(s.log.Named("exchange")).
		WithRand(s.cipher.Rand()).
		Server(s.key).Run(ctx)
	if err != nil {
		return crypto.AuthKey{}, err
	}

	return r.Key, nil
}
