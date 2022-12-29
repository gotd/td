package peers

import (
	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
)

func userPeerID(id int64) (r constant.TDLibPeerID) {
	r.User(id)
	return r
}

func chatPeerID(id int64) (r constant.TDLibPeerID) {
	r.Chat(id)
	return r
}

func channelPeerID(id int64) (r constant.TDLibPeerID) {
	r.Channel(id)
	return r
}

func peerIDFromPeerClass(p tg.PeerClass) constant.TDLibPeerID {
	switch p := p.(type) {
	case *tg.PeerUser:
		return userPeerID(p.UserID)
	case *tg.PeerChat:
		return chatPeerID(p.ChatID)
	case *tg.PeerChannel:
		return channelPeerID(p.ChannelID)
	}
	return 0
}

func (m *Manager) updated(ids ...constant.TDLibPeerID) {
	m.needUpdateMux.Lock()
	defer m.needUpdateMux.Unlock()

	m.needUpdate.delete(ids...)
}

func (m *Manager) updatedFull(id constant.TDLibPeerID) {
	m.needUpdateMux.Lock()
	defer m.needUpdateMux.Unlock()

	m.needUpdateFull.delete(id)
}

func (m *Manager) needsUpdate(id constant.TDLibPeerID) bool {
	m.needUpdateMux.Lock()
	defer m.needUpdateMux.Unlock()

	return m.needUpdate.has(id)
}

func (m *Manager) needsUpdateFull(id constant.TDLibPeerID) bool {
	m.needUpdateMux.Lock()
	defer m.needUpdateMux.Unlock()

	return m.needUpdateFull.has(id)
}

func (m *Manager) applyUpdates(updates []tg.UpdateClass) {
	// TODO(tdakkota): support partial updates in storage

	m.needUpdateMux.Lock()
	defer m.needUpdateMux.Unlock()

	appendBoth := func(p ...constant.TDLibPeerID) {
		m.needUpdate.add(p...)
		m.needUpdateFull.add(p...)
	}

	for _, update := range updates {
		switch update := update.(type) {
		case *tg.UpdateChatParticipants:
			p, ok := update.Participants.(*tg.ChatParticipants)
			if ok {
				appendBoth(chatPeerID(p.ChatID))
			}
		case *tg.UpdateUserStatus:
			m.needUpdate.add(userPeerID(update.UserID))
		case *tg.UpdateUserName:
			m.needUpdate.add(userPeerID(update.UserID))
		case *tg.UpdateUser:
			m.needUpdate.add(userPeerID(update.UserID))
		case *tg.UpdateUserPhone:
			m.needUpdate.add(userPeerID(update.UserID))
		case *tg.UpdateChatParticipantAdd:
			appendBoth(userPeerID(update.UserID), chatPeerID(update.ChatID))
		case *tg.UpdateChatParticipantDelete:
			appendBoth(userPeerID(update.UserID), chatPeerID(update.ChatID))
		case *tg.UpdateNotifySettings:
			if p, ok := update.Peer.(*tg.NotifyPeer); ok {
				m.needUpdate.add(peerIDFromPeerClass(p.Peer))
			}
		case *tg.UpdateChannel:
			m.needUpdate.add(channelPeerID(update.ChannelID))
		case *tg.UpdateChatParticipantAdmin:
			m.needUpdate.add(userPeerID(update.UserID), chatPeerID(update.ChatID))
		case *tg.UpdateChatDefaultBannedRights:
			appendBoth(peerIDFromPeerClass(update.Peer))
		case *tg.UpdatePeerSettings:
			m.needUpdate.add(peerIDFromPeerClass(update.Peer))
		case *tg.UpdatePeerBlocked:
		case *tg.UpdateChat:
			m.needUpdate.add(chatPeerID(update.ChatID))
		case *tg.UpdatePeerHistoryTTL:
			m.needUpdate.add(peerIDFromPeerClass(update.Peer))
		case *tg.UpdateChatParticipant:
			m.needUpdate.add(userPeerID(update.UserID), chatPeerID(update.ChatID))
		case *tg.UpdateChannelParticipant:
			m.needUpdate.add(userPeerID(update.UserID), channelPeerID(update.ChannelID))
		case *tg.UpdateBotStopped:
		case *tg.UpdateBotCommands:
			m.needUpdate.add(userPeerID(update.BotID))
		case *tg.UpdatePendingJoinRequests:
			m.needUpdate.add(peerIDFromPeerClass(update.Peer))
		case *tg.UpdateBotChatInviteRequester:
		}
	}
}
