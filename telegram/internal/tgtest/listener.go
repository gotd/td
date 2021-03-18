package tgtest

import (
	"context"
	"net"
)

func newLocalListener(ctx context.Context) (net.Listener, error) {
	conf := net.ListenConfig{}
	l, err := conf.Listen(ctx, "tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	return l, nil
}
