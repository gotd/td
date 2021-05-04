package dcs

import (
	"net"

	"github.com/gotd/td/transport"
)

type protocol interface {
	Handshake(conn net.Conn) (transport.Conn, error)
}
