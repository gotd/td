package mtproto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/testutil"
)

func TestCheckMessageID(t *testing.T) {
	now := testutil.Date()
	t.Run("Good", func(t *testing.T) {
		for _, good := range []proto.MessageID{
			proto.NewMessageID(now, proto.MessageFromServer),
			proto.NewMessageID(now, proto.MessageServerResponse),
			proto.NewMessageID(now.Add(time.Second*29), proto.MessageFromServer),
			proto.NewMessageID(now.Add(-time.Second*299), proto.MessageFromServer),
		} {
			t.Run(good.String(), func(t *testing.T) {
				require.NoError(t, checkMessageID(now, int64(good)))
			})
		}
	})
	t.Run("Bad", func(t *testing.T) {
		for _, bad := range []proto.MessageID{
			proto.NewMessageID(now, proto.MessageFromClient),
			proto.NewMessageID(now.Add(time.Second*31), proto.MessageFromServer),
			proto.NewMessageID(now.Add(-time.Second*301), proto.MessageFromServer),
			proto.NewMessageID(time.Time{}, proto.MessageFromServer),
			proto.NewMessageID(now.AddDate(-10, 0, 0), proto.MessageServerResponse),
			proto.NewMessageID(time.Time{}, proto.MessageFromClient),
		} {
			t.Run(bad.String(), func(t *testing.T) {
				require.ErrorIs(t, checkMessageID(now, int64(bad)), errRejected)
			})
		}
	})
}
