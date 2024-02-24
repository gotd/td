package codec

import (
	"fmt"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
)

const (
	// CodeAuthKeyNotFound means that specified auth key ID cannot be found by the DC.
	// Also, may be returned during key exchange.
	CodeAuthKeyNotFound = 404

	// CodeWrongDC means that current DC is wrong.
	// Usually returned by server when key exchange sends wrong DC ID.
	CodeWrongDC = 444

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
	case CodeWrongDC:
		return "wrong DC"
	default:
		return fmt.Sprintf("protocol error %d", p.Code)
	}
}

// Can be bigger that 1mb.
//
// See https://github.com/gotd/td/issues/412
//
// See https://github.com/tdlib/td/blob/550ccc8d9bbbe9cff1dc618aef5764d2cbd2cd91/td/mtproto/TcpTransport.cpp#L53
const maxMessageSize = 1 << 24 // 16 MB

func checkOutgoingMessage(b *bin.Buffer) error {
	length := b.Len()
	if length > maxMessageSize || length == 0 {
		return invalidMsgLenErr{n: length}
	}
	return nil
}

func checkAlign(b *bin.Buffer, n int) error {
	length := b.Len()
	if length%n != 0 {
		return alignedPayloadExpectedErr{expected: n}
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

type alignedPayloadExpectedErr struct {
	expected int
}

func (e alignedPayloadExpectedErr) Error() string {
	return fmt.Sprintf("payload is not aligned, expected align by %d", e.expected)
}

func (e alignedPayloadExpectedErr) Is(err error) bool {
	_, ok := err.(alignedPayloadExpectedErr)
	return ok
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
