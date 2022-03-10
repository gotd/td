package members

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

func convertInputUserToInputPeer(p tg.InputUserClass) tg.InputPeerClass {
	switch p := p.(type) {
	case *tg.InputUserSelf:
		return &tg.InputPeerSelf{}
	case *tg.InputUser:
		return &tg.InputPeerUser{
			UserID:     p.UserID,
			AccessHash: p.AccessHash,
		}
	case *tg.InputUserFromMessage:
		return &tg.InputPeerUserFromMessage{
			Peer:   p.Peer,
			MsgID:  p.MsgID,
			UserID: p.UserID,
		}
	default:
		return nil
	}
}

func editDefaultRights(ctx context.Context, api *tg.Client, p tg.InputPeerClass, rights MemberRights) error {
	if _, err := api.MessagesEditChatDefaultBannedRights(ctx, &tg.MessagesEditChatDefaultBannedRightsRequest{
		Peer:         p,
		BannedRights: rights.IntoChatBannedRights(),
	}); err != nil {
		return errors.Wrap(err, "edit default rights")
	}
	return nil
}
