package cluster

import (
	"context"
	"net"

	"github.com/go-faster/errors"
)

func newLocalListener(ctx context.Context) (net.Listener, error) {
	cfg := net.ListenConfig{}
	l, err := cfg.Listen(ctx, "tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, errors.Wrap(err, "listen")
	}
	return l, nil
}
