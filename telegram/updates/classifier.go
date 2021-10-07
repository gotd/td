package updates

import (
	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
)

func isCommonPtsUpdate(u tg.UpdateClass) (pts, ptsCount int, ok bool) {
	switch u := u.(type) {
	case *tg.UpdateNewMessage:
		return u.Pts, u.PtsCount, true
	case *tg.UpdateDeleteMessages:
		return u.Pts, u.PtsCount, true
	case *tg.UpdateReadHistoryInbox:
		return u.Pts, u.PtsCount, true
	case *tg.UpdateReadHistoryOutbox:
		return u.Pts, u.PtsCount, true
	case *tg.UpdateWebPage:
		return u.Pts, u.PtsCount, true
	case *tg.UpdateReadMessagesContents:
		return u.Pts, u.PtsCount, true
	case *tg.UpdateEditMessage:
		return u.Pts, u.PtsCount, true
	case *tg.UpdateFolderPeers:
		return u.Pts, u.PtsCount, true
	case *tg.UpdatePinnedMessages:
		return u.Pts, u.PtsCount, true
	}

	return
}

func isCommonQtsUpdate(u tg.UpdateClass) (qts int, ok bool) {
	switch u := u.(type) {
	case *tg.UpdateNewEncryptedMessage:
		return u.Qts, true
	case *tg.UpdateChatParticipant:
		return u.Qts, true
	case *tg.UpdateBotStopped:
		return u.Qts, true
	}

	return
}

func isChannelPtsUpdate(u tg.UpdateClass) (channelID int64, pts, ptsCount int, ok bool, err error) {
	switch u := u.(type) {
	case *tg.UpdateNewChannelMessage:
		channelID, err = extractChannelID(u.Message)
		return channelID, u.Pts, u.PtsCount, true, err
	case *tg.UpdateReadChannelInbox:
		return u.ChannelID, u.Pts, 0, true, nil
	case *tg.UpdateDeleteChannelMessages:
		return u.ChannelID, u.Pts, u.PtsCount, true, nil
	case *tg.UpdateEditChannelMessage:
		channelID, err = extractChannelID(u.Message)
		return channelID, u.Pts, u.PtsCount, true, err
	case *tg.UpdateChannelWebPage:
		return u.ChannelID, u.Pts, u.PtsCount, true, nil
	case *tg.UpdatePinnedChannelMessages:
		return u.ChannelID, u.Pts, u.PtsCount, true, nil
	case *tg.UpdateChannelParticipant:
		// TODO: ptsCount 1?
		return u.ChannelID, u.Qts, 0, true, nil
	}

	return
}

func extractChannelID(msg tg.MessageClass) (int64, error) {
	switch msg := msg.(type) {
	case *tg.Message:
		if c, ok := msg.PeerID.(*tg.PeerChannel); ok {
			return c.ChannelID, nil
		}

		return 0, xerrors.Errorf("unexpected tg.Message peer type: %T", msg.PeerID)
	case *tg.MessageEmpty:
		peer, ok := msg.GetPeerID()
		if !ok {
			return 0, xerrors.New("tg.MessageEmpty have no peerID field")
		}

		if c, ok := peer.(*tg.PeerChannel); ok {
			return c.ChannelID, nil
		}

		return 0, xerrors.Errorf("unexpected tg.MessageEmpty peer type: %T", peer)
	case *tg.MessageService:
		if c, ok := msg.PeerID.(*tg.PeerChannel); ok {
			return c.ChannelID, nil
		}

		return 0, xerrors.Errorf("unexpected tg.MessageService peer type: %T", msg.PeerID)
	default:
		return 0, xerrors.Errorf("unexpected tg.MessageClass type: %T", msg)
	}
}
