package wsutil

import (
	"context"
	"io"
	"math"
	"net"
	"sync"
	"time"

	"go.uber.org/multierr"
	"nhooyr.io/websocket"
)

var _ net.Conn = (*wsConn)(nil)

type wsConn struct {
	conn       *websocket.Conn
	localAddr  net.Addr
	remoteAddr net.Addr

	writeTimer   *time.Timer
	writeContext context.Context

	readTimer   *time.Timer
	readContext context.Context

	readMu sync.Mutex
	reader io.Reader

	onceClose sync.Once
	closeErr  error
}

// NetConn creates opaque wrapper net.Conn for websocket.Conn.
func NetConn(c *websocket.Conn, local, remote net.Addr) net.Conn {
	nc := &wsConn{
		conn:       c,
		localAddr:  local,
		remoteAddr: remote,
	}

	ctx := context.Background()

	var cancel context.CancelFunc
	nc.writeContext, cancel = context.WithCancel(ctx)
	nc.writeTimer = time.AfterFunc(math.MaxInt64, cancel)
	if !nc.writeTimer.Stop() {
		<-nc.writeTimer.C
	}

	nc.readContext, cancel = context.WithCancel(ctx)
	nc.readTimer = time.AfterFunc(math.MaxInt64, cancel)
	if !nc.readTimer.Stop() {
		<-nc.readTimer.C
	}

	return nc
}

func (w *wsConn) Write(b []byte) (int, error) {
	err := w.conn.Write(w.writeContext, websocket.MessageBinary, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (w *wsConn) Read(b []byte) (int, error) {
	w.readMu.Lock()
	defer w.readMu.Unlock()

	for {
		if w.reader == nil {
			// Advance to next message.
			var err error
			_, w.reader, err = w.conn.Reader(w.readContext)
			if err != nil {
				return 0, err
			}
		}
		n, err := w.reader.Read(b)
		if err == io.EOF {
			// At end of message.
			w.reader = nil
			if n > 0 {
				return n, nil
			}

			// No data read, continue to next message.
			continue
		}
		return n, err
	}
}

func (w *wsConn) Close() error {
	w.onceClose.Do(func() {
		w.writeTimer.Stop()
		w.readTimer.Stop()
		w.closeErr = w.conn.Close(websocket.StatusNormalClosure, "")
	})
	return w.closeErr
}

func (w *wsConn) LocalAddr() net.Addr {
	return w.localAddr
}

func (w *wsConn) RemoteAddr() net.Addr {
	return w.remoteAddr
}

func (w *wsConn) SetDeadline(t time.Time) error {
	return multierr.Append(w.SetWriteDeadline(t), w.SetReadDeadline(t))
}

func (w *wsConn) SetWriteDeadline(t time.Time) error {
	if t.IsZero() {
		w.writeTimer.Stop()
	} else {
		w.writeTimer.Reset(time.Until(t))
	}
	return nil
}

func (w *wsConn) SetReadDeadline(t time.Time) error {
	if t.IsZero() {
		w.readTimer.Stop()
	} else {
		w.readTimer.Reset(time.Until(t))
	}
	return nil
}
