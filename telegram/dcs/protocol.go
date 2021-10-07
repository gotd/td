package dcs

import (
	"net"

	"github.com/nnqq/td/transport"
)

type protocol interface {
	Handshake(conn net.Conn) (transport.Conn, error)
}
