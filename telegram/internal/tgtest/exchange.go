package tgtest

import (
	"context"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/telegram/internal/exchange"
	"github.com/gotd/td/transport"
)

// nolint:gocognit,gocyclo // TODO(tdakkota): simplify
func (s *Server) exchange(ctx context.Context, read *bin.Buffer, conn transport.Conn) (crypto.AuthKeyWithID, error) {
	cfg := exchange.NewConfig(s.clock, s.cipher.Rand(), conn, s.log.Named("exchange"))
	r, err := exchange.NewServerExchange(cfg, s.key).Run(ctx, read)
	if err != nil {
		return crypto.AuthKeyWithID{}, err
	}

	return r.Key, nil
}
