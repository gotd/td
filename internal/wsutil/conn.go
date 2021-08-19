package wsutil

import (
	"context"
	"net"

	"nhooyr.io/websocket"
)

// NetConn creates opaque wrapper net.Conn for websocket.Conn.
func NetConn(c *websocket.Conn) net.Conn {
	return websocket.NetConn(context.Background(), c, websocket.MessageBinary)
}
