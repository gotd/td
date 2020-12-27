package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

type dataHandler struct {
	log           *zap.Logger
	updateHandler UpdateHandler
	ctx           context.Context
}

func newDataHandler(ctx context.Context, logger *zap.Logger, uh UpdateHandler) *dataHandler {
	return &dataHandler{
		log:           logger,
		updateHandler: uh,
		ctx:           ctx,
	}
}

// mtproto.Handler
func (d *dataHandler) handleData(b *bin.Buffer) error {
	id, err := b.PeekID()
	if err != nil {
		return xerrors.Errorf("peek id: %w", err)
	}

	switch id {
	// Handling all types of tg.UpdatesClass.
	case tg.UpdatesTooLongTypeID,
		tg.UpdateShortMessageTypeID,
		tg.UpdateShortChatMessageTypeID,
		tg.UpdateShortTypeID,
		tg.UpdatesCombinedTypeID,
		tg.UpdatesTypeID,
		tg.UpdateShortSentMessageTypeID:
		updates, err := tg.DecodeUpdates(b)
		if err != nil {
			return xerrors.Errorf("decode updates: %w", err)
		}

		d.log.Info("Processing updates...")
		return d.processUpdates(updates)
	default:
		return d.handleUnknown(b)
	}
}

func (d *dataHandler) processUpdates(updates tg.UpdatesClass) error {
	switch u := updates.(type) {
	case *tg.Updates:
		return d.updateHandler(d.ctx, u)
	case *tg.UpdateShort:
		// TODO(ernado): separate handler
		return d.updateHandler(d.ctx, &tg.Updates{
			Date: u.Date,
			Updates: []tg.UpdateClass{
				u.Update,
			},
		})
	// TODO(ernado): handle UpdatesTooLong
	// TODO(ernado): handle UpdateShortMessage
	// TODO(ernado): handle UpdateShortChatMessage
	// TODO(ernado): handle UpdatesCombined
	// TODO(ernado): handle UpdateShortSentMessage
	default:
		d.log.With(zap.String("update_type", fmt.Sprintf("%T", u))).Warn("Ignoring update")
	}
	return nil
}

func (d *dataHandler) handleUnknown(b *bin.Buffer) error {
	// Can't process unknown type.
	id, err := b.PeekID()
	if err != nil {
		return err
	}
	d.log.With(
		zap.String("type_id", fmt.Sprintf("0x%x", id)),
	).Warn("Unknown type id")

	return nil
}
