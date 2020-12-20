package tgtest

import (
	"fmt"
	"net"

	"golang.org/x/net/nettest"
)

func newLocalListener() net.Listener {
	l, err := nettest.NewLocalListener("tcp")
	if err != nil {
		panic(fmt.Sprintf("tgtest: failed to listen on a port: %v", err))
	}
	return l
}
