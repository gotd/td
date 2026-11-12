package mtproto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/neo"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/testutil"
)

func TestServerTimeObserve(t *testing.T) {
	a := require.New(t)
	clk := neo.NewTime(testutil.Date())
	st := NewServerTime(clk, 3*time.Second) // seeded, e.g. persisted
	a.Equal(3*time.Second, st.Offset())
	a.False(st.Known())

	// The first observation replaces the seed, also downwards.
	a.True(st.observe(clk.Now().Add(-10 * time.Minute)))
	a.True(st.Known())
	a.Equal(-10*time.Minute, st.Offset())
	a.Equal(clk.Now().Add(-10*time.Minute), st.Now())

	// Later ones only raise it (messages arrive after they were created).
	a.False(st.observe(clk.Now().Add(-11 * time.Minute)))
	a.Equal(-10*time.Minute, st.Offset())
	a.False(st.observe(clk.Now().Add(-10*time.Minute + 50*time.Microsecond)))
	a.True(st.observe(clk.Now().Add(-9 * time.Minute)))
	a.Equal(-9*time.Minute, st.Offset())

	// reset lowers it.
	st.reset(clk.Now().Add(-12 * time.Minute))
	a.Equal(-12*time.Minute, st.Offset())
}

// bad_msg_notification 17 (msg_id too high) resets the offset from the
// notification's own msg_id, not from bad_msg_id.
func TestHandleBadMsgTooHighResetsServerTime(t *testing.T) {
	a := require.New(t)
	clk := neo.NewTime(testutil.Date())
	c := newTestClient(func(int64, int32, bin.Encoder) (bin.Encoder, error) { return nil, nil })
	c.serverTime = NewServerTime(clk, 0)
	a.True(c.serverTime.observe(clk.Now().Add(time.Minute))) // offset +1m

	notifID := proto.NewMessageID(clk.Now().Add(-2*time.Minute), proto.MessageFromServer)
	var b bin.Buffer
	a.NoError(b.Encode(&mt.BadMsgNotification{
		BadMsgID:  int64(proto.NewMessageID(clk.Now().Add(time.Hour), proto.MessageFromClient)),
		ErrorCode: codeMessageIDTooHigh,
	}))
	a.NoError(c.handleBadMsg(int64(notifID), &b))
	a.InDelta(float64(-2*time.Minute), float64(c.serverTime.Offset()), float64(time.Millisecond))
}

// A request refused for its msg_id time is sent again with a new msg_id.
func TestInvokeResendsAfterBadMsgTime(t *testing.T) {
	for _, code := range []int{codeMessageIDTooLow, codeMessageIDTooHigh, codeMessageTooOld} {
		a := require.New(t)
		var ids []int64
		c := newTestClient(func(msgID int64, _ int32, in bin.Encoder) (bin.Encoder, error) {
			ids = append(ids, msgID)
			if len(ids) == 1 {
				return nil, &badMessageError{Code: code}
			}
			return &mt.Pong{MsgID: msgID, PingID: 1}, nil
		})
		var pong mt.Pong
		a.NoError(c.Invoke(t.Context(), &mt.PingRequest{PingID: 1}, &pong), "code %d", code)
		a.Len(ids, 2)
		a.NotEqual(ids[0], ids[1], "new msg_id")
	}
}

func TestInvokeBadMsgTimeResendLimit(t *testing.T) {
	a := require.New(t)
	calls := 0
	c := newTestClient(func(int64, int32, bin.Encoder) (bin.Encoder, error) {
		calls++
		return nil, &badMessageError{Code: codeMessageIDTooLow}
	})
	err := c.Invoke(t.Context(), &mt.PingRequest{PingID: 1}, &mt.Pong{})
	var bad *badMessageError
	a.ErrorAs(err, &bad)
	a.Equal(1+maxTimeResends, calls)
}
