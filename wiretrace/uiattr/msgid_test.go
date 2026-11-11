package uiattr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMsgIDLabel(t *testing.T) {
	for _, tt := range []struct {
		id   int64
		want string
	}{
		{0, "unread"},
		{1, "1"},
		{4614049, "4614049"},
		{72057594037927935, "newest"}, // ServerMaxMsgId - 1, seen on production
		{ServerMaxMsgID + 5, "scheduled(5)"},
		{ScheduledMaxMsgID + 2, "shortcut(2)"},
		{StartClientMsgID + 7, "local(7)"},
		{StartStoryMsgID + 3, "story(3)"},
		{SpecialMsgIDShift + 1, "end"},
		{SpecialMsgIDShift + 2, "top"},
		{SpecialMsgIDShift + 4, "start_bot"},
		{SpecialMsgIDShift + 6, "choose_messages"},
		{SpecialMsgIDShift + 3, "special(-288230372930486267)"},
		{-5, "special(-5)"},
	} {
		require.Equal(t, tt.want, MsgIDLabel(tt.id), "id %d", tt.id)
	}
}
