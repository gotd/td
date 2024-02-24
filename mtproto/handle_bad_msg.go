package mtproto

import (
	"fmt"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
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
	description := map[int]string{
		codeMessageIDTooLow:     "msg_id too low",
		codeMessageIDTooHigh:    "msg_id too high",
		codeIncorrectServerSalt: "incorrect server salt",

		18: "incorrect two lower order msg_id bits",
		19: "container msg_id is the same as msg_id of a previously received message",
		20: "message too old",
		32: "msg_seqno too low",
		33: "msg_seqno too high",
		34: "even msg_seqno expected, but odd received",
		35: "odd msg_seqno expected, but even received",
	}[c.Code]
	if description == "" {
		return fmt.Sprintf("bad msg error code %d", c.Code)
	}
	return description
}

func (c *Conn) handleBadMsg(b *bin.Buffer) error {
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
		return errors.Errorf("unknown type id 0x%d", id)
	}
}
