package codec

import (
	"errors"
	"fmt"

	"github.com/gotd/td/bin"
)

const (
	// CodeAuthKeyNotFound means that specified auth key ID cannot be found by the DC.
	CodeAuthKeyNotFound = 404

	// CodeTransportFlood means that too many transport connections are
	// established to the same IP in a too short lapse of time, or if any
	// of the container/service message limits are reached.
	CodeTransportFlood = 429
)

// ProtocolErr represents protocol level error.
type ProtocolErr struct {
	Code int32
}

func (p ProtocolErr) Error() string {
	switch p.Code {
	case CodeAuthKeyNotFound:
		return "auth key not found"
	case CodeTransportFlood:
		return "transport flood"
	default:
		return fmt.Sprintf("protocol error %d", p.Code)
	}
}

var errMessageMustBePadded = errors.New("message must be padded by 4")

const maxMessageSize = 1024 * 1024 // 1mb

func checkOutgoingMessage(b *bin.Buffer) error {
	length := b.Len()
	if length > maxMessageSize {
		return invalidMsgLenErr{n: length}
	}

	if length%4 != 0 {
		return errMessageMustBePadded
	}

	return nil
}

func checkProtocolError(b *bin.Buffer) error {
	if b.Len() != bin.Word {
		return nil
	}
	code, err := b.Int32()
	if err != nil {
		return err
	}
	return &ProtocolErr{Code: -code}
}

type invalidMsgLenErr struct {
	n int
}

func (e invalidMsgLenErr) Error() string {
	return fmt.Sprintf("invalid message length %d", e.n)
}

func (e invalidMsgLenErr) Is(err error) bool {
	_, ok := err.(invalidMsgLenErr)
	return ok
}

// ErrProtocolHeaderMismatch means that received protocol header
// is mismatched with expected.
var ErrProtocolHeaderMismatch = errors.New("protocol header mismatch")
