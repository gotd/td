package updates

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

func validatePts(pts, ptsCount int) error {
	if pts < 0 {
		return errors.Errorf("invalid pts value: %d", pts)
	}

	if ptsCount < 0 {
		return errors.Errorf("invalid ptsCount value: %d", ptsCount)
	}

	return nil
}

func validateQts(qts int) error {
	if qts < 0 {
		return errors.Errorf("invalid qts value: %d", qts)
	}

	return nil
}

func validateSeq(seq, seqStart int) error {
	if seq < 0 {
		return errors.Errorf("invalid seq value: %d", seq)
	}

	if seqStart < 0 {
		return errors.Errorf("invalid seqStart value: %d", seq)
	}

	return nil
}

func getDialogPts(dialog tg.DialogClass) (int, error) {
	d, ok := dialog.(*tg.Dialog)
	if !ok {
		return 0, errors.Errorf("unexpected dialog type: %T", dialog)
	}

	pts, ok := d.GetPts()
	if !ok {
		return 0, errors.New("dialog has no pts field")
	}

	return pts, nil
}

func msgsToUpdates(msgs []tg.MessageClass, channel bool) []tg.UpdateClass {
	updates := make([]tg.UpdateClass, 0, len(msgs))
	for _, msg := range msgs {
		if channel {
			updates = append(updates, &tg.UpdateNewChannelMessage{
				Message:  msg,
				Pts:      -1,
				PtsCount: -1,
			})
			continue
		}

		updates = append(updates, &tg.UpdateNewMessage{
			Message:  msg,
			Pts:      -1,
			PtsCount: -1,
		})
	}

	return updates
}

func encryptedMsgsToUpdates(msgs []tg.EncryptedMessageClass) []tg.UpdateClass {
	updates := make([]tg.UpdateClass, 0, len(msgs))
	for _, msg := range msgs {
		updates = append(updates, &tg.UpdateNewEncryptedMessage{
			Message: msg,
			Qts:     -1,
		})
	}

	return updates
}
