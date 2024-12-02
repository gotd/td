package wsutil

import (
	"context"
	"net"

	"github.com/coder/websocket"
)

// NetConn creates opaque wrapper net.Conn for websocket.Conn.
func NetConn(c *websocket.Conn) net.Conn {
	return websocket.NetConn(context.Background(), c, websocket.MessageBinary)
}
