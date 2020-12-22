package telegram

import (
	"fmt"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
)

type badMessageError struct {
	Code    int
	NewSalt int64
}

const (
	codeMessageIDTooLow     = 16
	codeMessageIDTooHigh    = 17
	codeIncorrectServerSalt = 48
)

func (c badMessageError) Error() string {
	switch c.Code {
	case codeMessageIDTooLow:
		return "msg_id too low"
	case codeMessageIDTooHigh:
		return "msg_id too high"
	case 18:
		return "incorrect two lower order msg_id bits"
	case 19:
		return "container msg_id is the same as msg_id of a previously received message"
	case 20:
		return "message too old"
	case 32:
		// the server has already received a message with a lower msg_id
		// but with either a higher or an equal and odd seqno
		return "msg_seqno too low"
	case 33:
		return "msg_seqno too high"
	case 34:
		return "even msg_seqno expected, but odd received"
	case 35:
		return "odd msg_seqno expected, but even received"
	case codeIncorrectServerSalt:
		return "incorrect server salt"
	default:
		return fmt.Sprintf("bad msg error code %d", c.Code)
	}
}

func (c *Client) handleBadMsg(b *bin.Buffer) error {
	id, err := b.PeekID()
	if err != nil {
		return err
	}
	switch id {
	case mt.BadMsgNotificationTypeID:
		var bad mt.BadMsgNotification
		if err := bad.Decode(b); err != nil {
			return err
		}

		c.rpc.NotifyError(bad.BadMsgID, &badMessageError{Code: bad.ErrorCode})
		return nil
	case mt.BadServerSaltTypeID:
		var bad mt.BadServerSalt
		if err := bad.Decode(b); err != nil {
			return err
		}

		c.rpc.NotifyError(bad.BadMsgID, &badMessageError{Code: bad.ErrorCode, NewSalt: bad.NewServerSalt})
		return nil
	default:
		return xerrors.Errorf("unknown type id 0x%d", id)
	}
}
