package updates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func TestMessageUserIDs(t *testing.T) {
	const selfID = 999

	full := func() *tg.Message {
		m := &tg.Message{
			ID:     1,
			PeerID: &tg.PeerUser{UserID: 111},
			FromID: &tg.PeerUser{UserID: 222},
		}
		var fwd tg.MessageFwdHeader
		fwd.SetFromID(&tg.PeerUser{UserID: 333})
		fwd.SetSavedFromPeer(&tg.PeerUser{UserID: 444})
		m.SetFwdFrom(fwd)
		m.SetViaBotID(555)
		m.SetEntities([]tg.MessageEntityClass{
			&tg.MessageEntityMentionName{UserID: 666},
			&tg.MessageEntityBold{},
		})
		return m
	}

	tests := []struct {
		name string
		msg  *tg.Message
		want []int64
	}{
		{
			name: "sender only",
			msg:  &tg.Message{PeerID: &tg.PeerUser{UserID: 111}},
			want: []int64{111},
		},
		{
			name: "self and zero excluded",
			msg:  &tg.Message{PeerID: &tg.PeerUser{UserID: selfID}, FromID: &tg.PeerUser{UserID: 0}},
			want: []int64{},
		},
		{
			name: "non-user peers ignored",
			msg:  &tg.Message{PeerID: &tg.PeerChannel{ChannelID: 42}, FromID: &tg.PeerChat{ChatID: 7}},
			want: []int64{},
		},
		{
			name: "peer + sender + fwd + saved-from + via_bot + mention",
			msg:  full(),
			want: []int64{111, 222, 333, 444, 555, 666},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, messageUserIDs(selfID, tt.msg))
		})
	}
}

func TestMessageUpdatesPeersKnown(t *testing.T) {
	ctx := context.Background()
	handler := telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
		return nil
	})
	s := newShortTestState(t, &countDiffAPI{}, handler)
	require.NoError(t, s.userHasher.SetUserAccessHash(ctx, s.selfID, 111, 7777))

	knownMsg := func(senderID int64) *tg.Message {
		return &tg.Message{ID: 1, PeerID: &tg.PeerUser{UserID: senderID}}
	}

	tests := []struct {
		name    string
		updates []tg.UpdateClass
		want    bool
	}{
		{
			name:    "non-message update is ignored",
			updates: []tg.UpdateClass{&tg.UpdateUserName{UserID: 222}},
			want:    true,
		},
		{
			name:    "non-*tg.Message body is ignored",
			updates: []tg.UpdateClass{&tg.UpdateNewMessage{Message: &tg.MessageService{ID: 1, PeerID: &tg.PeerUser{UserID: 222}}}},
			want:    true,
		},
		{
			name:    "new message with known sender",
			updates: []tg.UpdateClass{&tg.UpdateNewMessage{Message: knownMsg(111)}},
			want:    true,
		},
		{
			name:    "new message with unknown sender",
			updates: []tg.UpdateClass{&tg.UpdateNewMessage{Message: knownMsg(222)}},
			want:    false,
		},
		{
			name:    "edit message with unknown sender",
			updates: []tg.UpdateClass{&tg.UpdateEditMessage{Message: knownMsg(222)}},
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, s.messageUpdatesPeersKnown(ctx, tt.updates))
		})
	}
}
