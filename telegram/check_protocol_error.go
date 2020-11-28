package telegram

import (
	"fmt"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/internal/proto"
)

// ProtocolErr represents protocol level error.
type ProtocolErr struct {
	Code int32
}

func (p ProtocolErr) Error() string {
	switch p.Code {
	case proto.CodeAuthKeyNotFound:
		return "auth key not found"
	case proto.CodeTransportFlood:
		return "transport flood"
	default:
		return fmt.Sprintf("protocol error %d", p.Code)
	}
}

func (c *Client) checkProtocolError(b *bin.Buffer) error {
	if b.Len() != bin.Word {
		return nil
	}
	code, err := b.Int32()
	if err != nil {
		return err
	}
	return &ProtocolErr{Code: -code}
}
