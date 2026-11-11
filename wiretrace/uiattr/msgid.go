package uiattr

import "strconv"

// Message id ranges of tdesktop (Telegram/SourceFiles/data/data_msg_id.h).
// Server message ids are in (0, ServerMaxMsgID); the client uses the other
// ranges for local, story, scheduled and shortcut messages and for special
// "show at" values.
const (
	ServerMaxMsgID    = int64(1) << 56
	ScheduledMaxMsgID = ServerMaxMsgID + int64(1)<<32
	ShortcutMaxMsgID  = ScheduledMaxMsgID + int64(1)<<32

	StartClientMsgID  = 1 - int64(1)<<58
	EndClientMsgID    = StartClientMsgID + int64(1)<<31
	StartStoryMsgID   = EndClientMsgID + 1
	EndStoryMsgID     = StartStoryMsgID + int64(1)<<30
	SpecialMsgIDShift = EndStoryMsgID
)

var specialMsgIDs = map[int64]string{
	0:                     "unread",
	ServerMaxMsgID - 1:    "newest",
	SpecialMsgIDShift + 1: "end",
	SpecialMsgIDShift + 2: "top",
	SpecialMsgIDShift + 4: "start_bot",
	SpecialMsgIDShift + 5: "maybe_start_bot",
	SpecialMsgIDShift + 6: "choose_messages",
}

// MsgIDLabel formats a tdesktop message id: server ids as numbers, special
// values by name (ShowAtUnreadMsgId is "unread", ShowAtTheEndMsgId "end",
// ServerMaxMsgId-1, the "newest message" anchor, "newest"), and ids of the
// client ranges as local(N), story(N), scheduled(N) or shortcut(N).
func MsgIDLabel(id int64) string {
	if name, ok := specialMsgIDs[id]; ok {
		return name
	}
	switch {
	case id > 0 && id < ServerMaxMsgID:
		return strconv.FormatInt(id, 10)
	case id > ServerMaxMsgID && id <= ScheduledMaxMsgID:
		return "scheduled(" + strconv.FormatInt(id-ServerMaxMsgID, 10) + ")"
	case id > ScheduledMaxMsgID && id <= ShortcutMaxMsgID:
		return "shortcut(" + strconv.FormatInt(id-ScheduledMaxMsgID, 10) + ")"
	case id >= StartClientMsgID && id < EndClientMsgID:
		return "local(" + strconv.FormatInt(id-StartClientMsgID, 10) + ")"
	case id >= StartStoryMsgID && id < EndStoryMsgID:
		return "story(" + strconv.FormatInt(id-StartStoryMsgID, 10) + ")"
	}
	return "special(" + strconv.FormatInt(id, 10) + ")"
}
