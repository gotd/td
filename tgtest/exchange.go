package tgtest

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/internal/exchange"
	"github.com/nnqq/td/internal/proto/codec"
	"github.com/nnqq/td/transport"
)

type exchangeConn struct {
	transport.Conn
}

func (e exchangeConn) Recv(ctx context.Context, b *bin.Buffer) error {
	for {
		if err := e.Conn.Recv(ctx, b); err != nil {
			return err
		}

		var authKeyID [8]byte
		if err := b.PeekN(authKeyID[:], len(authKeyID)); err != nil {
			return xerrors.Errorf("peek id: %w", err)
		}
		if authKeyID != [8]byte{} {
			// TODO(tdakkota): what if client send registered auth key during key exchange?
			buf := bin.Buffer{}
			buf.PutInt32(-codec.CodeAuthKeyNotFound)

			if err := e.Conn.Send(ctx, &buf); err != nil {
				return xerrors.Errorf("send: %w", err)
			}

			continue
		}

		return nil
	}
}

// exchange starts MTProto key exchange.
func (s *Server) exchange(ctx context.Context, conn transport.Conn) (crypto.AuthKey, error) {
	r, err := exchange.NewExchanger(conn, s.dcID).
		WithClock(s.clock).
		WithLogger(s.log.Named("exchange")).
		WithRand(s.cipher.Rand()).
		Server(s.key).Run(ctx)
	if err != nil {
		return crypto.AuthKey{}, err
	}

	return r.Key, nil
}
