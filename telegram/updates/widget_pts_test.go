package updates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

// updateMessageWidgetData carries pts/pts_count of the common (non-channel)
// sequence, like updatePinnedMessages; tdlib UpdatesManager::is_pts_update
// lists it among common pts updates even though it has a Peer field.
func TestMessageWidgetDataIsCommonPts(t *testing.T) {
	u := &tg.UpdateMessageWidgetData{
		Peer:     &tg.PeerUser{UserID: 1},
		MsgID:    10,
		Pts:      42,
		PtsCount: 1,
	}

	pts, ptsCount, ok := tg.IsPtsUpdate(u)
	require.True(t, ok)
	require.Equal(t, 42, pts)
	require.Equal(t, 1, ptsCount)

	_, _, _, ok, _ = tg.IsChannelPtsUpdate(u)
	require.False(t, ok)

	_, ok = tg.IsQtsUpdate(u)
	require.False(t, ok)
}

// TestMessageWidgetDataGap ensures updateMessageWidgetData takes part in common
// pts gap detection: it is postponed while a gap is open and both updates are
// delivered in order once the gap is filled.
func TestMessageWidgetDataGap(t *testing.T) {
	ctx := context.Background()
	h := &recordingHandler{}
	s := newAffectedTestState(t, h, 100)

	widget := &tg.UpdateMessageWidgetData{
		Peer:       &tg.PeerUser{UserID: 1},
		MsgID:      10,
		PublicData: "data",
		Pts:        102,
		PtsCount:   1,
	}
	edit := &tg.UpdateEditMessage{
		Message:  &tg.Message{ID: 10, PeerID: &tg.PeerUser{UserID: 1}},
		Pts:      101,
		PtsCount: 1,
	}

	// Gap: pts 101 is missing.
	require.NoError(t, s.handleUpdates(ctx, &tg.Updates{
		Updates: []tg.UpdateClass{widget},
		Date:    1,
	}))
	require.Empty(t, h.updates(), "widget data update must wait for the gap")
	require.Equal(t, 100, s.pts.State())

	require.NoError(t, s.handlePts(ctx, edit.Pts, edit.PtsCount, edit, entities{}))
	require.Equal(t, 102, s.pts.State(), "pts must advance past the widget data update")

	got := h.updates()
	require.Len(t, got, 2)
	require.IsType(t, &tg.UpdateEditMessage{}, got[0])
	require.IsType(t, &tg.UpdateMessageWidgetData{}, got[1])

	// Contiguous update is applied immediately.
	next := &tg.UpdateMessageWidgetData{
		Peer:     &tg.PeerUser{UserID: 1},
		MsgID:    10,
		Pts:      103,
		PtsCount: 1,
	}
	require.NoError(t, s.handleUpdates(ctx, &tg.Updates{Updates: []tg.UpdateClass{next}, Date: 2}))
	require.Equal(t, 103, s.pts.State())
	require.Len(t, h.updates(), 3)
}
