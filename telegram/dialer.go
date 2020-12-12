package telegram

import (
	"context"
	"net"
)

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}
