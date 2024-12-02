package transport

import (
	"bytes"
	"io"
	"net"
	"net/http"

	"github.com/coder/websocket"

	"github.com/gotd/td/mtproxy/obfuscated2"
	"github.com/gotd/td/proto/codec"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/wsutil"
)

type wsListener struct {
	addr   net.Addr
	ch     chan *wsServerConn
	closed *tdsync.Ready
}

// WebsocketListener creates new MTProto Websocket listener.
func WebsocketListener(addr net.Addr) (net.Listener, http.Handler) {
	l := wsListener{
		addr:   addr,
		ch:     make(chan *wsServerConn, 1),
		closed: tdsync.NewReady(),
	}
	return l, l
}

func (l wsListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols: []string{"binary"},
	})
	if err != nil {
		w.WriteHeader(400)
		return
	}
	defer func() {
		_ = wsConn.Close(websocket.StatusNormalClosure, "Close")
	}()

	conn := wsutil.NetConn(wsConn)
	rw, md, err := obfuscated2.Accept(conn, nil)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	var tag *bytes.Reader
	if md.Protocol[0] == codec.AbridgedClientStart[0] {
		// Abridged sends only byte for tag.
		tag = bytes.NewReader(md.Protocol[:1])
	} else {
		tag = bytes.NewReader(md.Protocol[:])
	}

	accepted := &wsServerConn{
		closed: *tdsync.NewReady(),
		// Add codec tag in the begin of stream to emulate TCP fully.
		// MTProto sends codec tag in plain TCP connections, but not in obfuscated2 (Websocket/MTProxy).
		reader: io.MultiReader(tag, rw),
		writer: rw,
		Conn:   conn,
	}

	reqCtx := r.Context().Done()
	closed := l.closed.Ready()

	// Pass connection to the Accept().
	select {
	case <-reqCtx:
		return
	case <-closed:
		return
	case l.ch <- accepted:
	}

	// Await close or shutdown.
	select {
	case <-reqCtx:
		return
	case <-closed:
		return
	case <-accepted.closed.Ready():
	}
}

func (l wsListener) Accept() (net.Conn, error) {
	r := l.closed.Ready()

	for {
		select {
		case <-r:
			return nil, net.ErrClosed
		case conn := <-l.ch:
			return conn, nil
		}
	}
}

func (l wsListener) Close() error {
	l.closed.Signal()
	return nil
}

func (l wsListener) Addr() net.Addr {
	return l.addr
}

type wsServerConn struct {
	closed tdsync.Ready
	reader io.Reader
	writer io.Writer
	net.Conn
}

func (c *wsServerConn) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

func (c *wsServerConn) Write(p []byte) (int, error) {
	return c.writer.Write(p)
}

func (c *wsServerConn) Close() error {
	c.closed.Signal()
	return nil
}
