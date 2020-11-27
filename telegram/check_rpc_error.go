package telegram

import (
	"golang.org/x/xerrors"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/internal/mt"
)

func (c Client) checkRPCError(b *bin.Buffer) error {
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
		return xerrors.Errorf("bad message (code %d)", bad.ErrorCode)
	case mt.BadServerSaltTypeID:
		var bad mt.BadServerSalt
		if err := bad.Decode(b); err != nil {
			return err
		}
		return xerrors.Errorf("bad salt (code %d)", bad.ErrorCode)
	default:
		return nil
	}
}
