package mtproto

import (
	"io"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"

	"golang.org/x/xerrors"
)

func TestShouldReconnect(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		for _, err := range []error{
			xerrors.Errorf("failed: %w", io.EOF),
			syscall.ECONNRESET,
			syscall.EPIPE,
			xerrors.Errorf("disconnect: %w", &os.SyscallError{
				Err:     syscall.EPIPE,
				Syscall: "read",
			}),
			xerrors.Errorf("net: %w", &net.OpError{
				Err: syscall.EPIPE,
				Op:  "write",
			}),
		} {
			assert.True(t, shouldReconnect(err), "should reconnect on %v", err)
		}
	})
	t.Run("False", func(t *testing.T) {
		for _, err := range []error{
			nil,
			io.ErrNoProgress,
			xerrors.Errorf("bad: %w", &os.SyscallError{
				Err:     syscall.EBADMSG,
				Syscall: "read",
			}),
			xerrors.Errorf("bad: %w", &os.SyscallError{
				Err:     syscall.EPIPE,
				Syscall: "bad",
			}),
		} {
			assert.False(t, shouldReconnect(err), "should not reconnect on %v", err)
		}
	})
}
