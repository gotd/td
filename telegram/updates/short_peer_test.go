package updates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestReferencedUserIDs(t *testing.T) {
	const selfID = 999

	withOptional := func() shortMessageOptional {
		u := &tg.UpdateShortMessage{UserID: 111}
		var fwd tg.MessageFwdHeader
		fwd.SetFromID(&tg.PeerUser{UserID: 222})
		fwd.SetSavedFromPeer(&tg.PeerUser{UserID: 666})
		u.SetFwdFrom(fwd)
		u.SetViaBotID(333)
		u.SetEntities([]tg.MessageEntityClass{
			&tg.MessageEntityMentionName{UserID: 444},
			&tg.MessageEntityBold{},
		})
		return u
	}

	tests := []struct {
		name    string
		primary int64
		opt     shortMessageOptional
		want    []int64
	}{
		{"sender only", 111, &tg.UpdateShortMessage{UserID: 111}, []int64{111}},
		{"self excluded", selfID, &tg.UpdateShortMessage{UserID: selfID}, []int64{}},
		{"fwd + saved-from + via_bot + mention", 111, withOptional(), []int64{111, 222, 666, 333, 444}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, referencedUserIDs(selfID, tt.primary, tt.opt))
		})
	}
}

func TestShortMessagePeersKnown(t *testing.T) {
	ctx := context.Background()
	const selfID = 999
	hasher := newMemUserAccessHasher()
	require.NoError(t, hasher.SetUserAccessHash(ctx, selfID, 111, 7777))
	s := &internalState{
		selfID:     selfID,
		userHasher: hasher,
	}

	require.True(t, s.shortMessagePeersKnown(ctx, &tg.UpdateShortMessage{UserID: 111}))
	require.False(t, s.shortMessagePeersKnown(ctx, &tg.UpdateShortMessage{UserID: 222}))
	// selfID is always known without being seeded.
	require.True(t, s.shortMessagePeersKnown(ctx, &tg.UpdateShortMessage{UserID: selfID}))
}

func TestUserPeersKnown(t *testing.T) {
	ctx := context.Background()
	const selfID = 999
	hasher := newMemUserAccessHasher()
	require.NoError(t, hasher.SetUserAccessHash(ctx, selfID, 111, 7777))
	s := &internalState{
		selfID:     selfID,
		userHasher: hasher,
	}

	require.True(t, s.userPeersKnown(ctx, nil), "no peers are trivially known")
	// selfID is skipped before the hasher lookup, so it counts as known.
	require.True(t, s.userPeersKnown(ctx, []int64{selfID, 111}))
	require.False(t, s.userPeersKnown(ctx, []int64{111, 222}), "an unknown peer makes the batch unknown")
}
