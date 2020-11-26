package telegram

import (
	"golang.org/x/xerrors"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/internal/proto"
)

func (c *Client) checkProtocolError(b *bin.Buffer) error {
	if b.Len() != bin.Word {
		return nil
	}
	code, err := b.Int32()
	if err != nil {
		return err
	}
	code *= -1
	switch code {
	case proto.CodeAuthKeyNotFound:
		return xerrors.New("protocol error: auth key not found")
	case proto.CodeTransportFlood:
		return xerrors.New("protocol error: transport flood")
	default:
		return xerrors.Errorf("protocol erorr: code %d", code)
	}
}
