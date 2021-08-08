package cluster

import (
	"context"
	"net"

	"golang.org/x/xerrors"
)

func newLocalListener(ctx context.Context) (net.Listener, error) {
	cfg := net.ListenConfig{}
	l, err := cfg.Listen(ctx, "tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, xerrors.Errorf("listen: %w", err)
	}
	return l, nil
}
