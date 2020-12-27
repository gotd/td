package mtproto

import (
	"errors"
	"io"
	"net"
	"os"
	"syscall"

	"golang.org/x/xerrors"
)

func (c *Client) reconnect() error {
	c.sessionCreated.Reset()
	c.log.Debug("Disconnected. Trying to re-connect")

	if err := c.connect(c.ctx); err != nil {
		return xerrors.Errorf("connect: %w", err)
	}

	return nil
}

// shouldReconnect returns true if err is caused by failed read on connection
// that was closed.
//
// E.g. write tcp 127.0.0.1:10->127.0.0.1:20: read: connection reset by peer
func shouldReconnect(err error) bool {
	isRW := func(op string) bool {
		switch op {
		case "read", "write":
			return true
		default:
			return false
		}
	}

	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) {
		return true
	}
	if err == syscall.ECONNRESET || err == syscall.EPIPE {
		return true
	}

	var sysErr *os.SyscallError
	if errors.As(err, &sysErr) {
		if !isRW(sysErr.Syscall) {
			return false
		}
		return shouldReconnect(sysErr.Err)
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if !isRW(opErr.Op) {
			return false
		}
		return shouldReconnect(opErr.Err)
	}

	return false
}
