package tgtest

import (
	"context"
	"fmt"
	"net"
)

func newLocalListener(ctx context.Context) net.Listener {
	conf := net.ListenConfig{}
	l, err := conf.Listen(ctx, "tcp4", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Sprintf("tgtest: failed to listen on a port: %v", err))
	}
	return l
}
