package mtproto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/neo"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/rpc"
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
// "msg_id too high" is handled by a new session instead, see
// TestInvokeMsgIDTooHighNewSession.
func TestInvokeResendsAfterBadMsgTime(t *testing.T) {
	for _, code := range []int{codeMessageIDTooLow, codeMessageTooOld} {
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

// "msg_id too high": the request is not sent again in the failed session
// (see failSession) but reported as retryable on the next connection.
func TestInvokeMsgIDTooHighRetryable(t *testing.T) {
	a := require.New(t)
	calls := 0
	c := newTestClient(func(int64, int32, bin.Encoder) (bin.Encoder, error) {
		calls++
		return nil, &badMessageError{Code: codeMessageIDTooHigh}
	})
	var pong mt.Pong
	err := c.Invoke(t.Context(), &mt.PingRequest{PingID: 1}, &pong)
	a.ErrorIs(err, rpc.ErrEngineClosed, "retryable on a new connection")
	a.Equal(1, calls, "no resend in the failed session")
}

// handleBadMsg with "msg_id too high" ends the session: the id generator is
// reset to the corrected time, Run is stopped (sessionFailed) and the next
// connection does not continue the session.
func TestHandleBadMsgTooHighFailsSession(t *testing.T) {
	a := require.New(t)
	clk := neo.NewTime(testutil.Date())
	skew := 10 * time.Minute
	gen := proto.NewMessageIDGen(func() time.Time { return clk.Now().Add(skew) })
	high := gen.New(proto.MessageFromClient) // sent with the clock ahead
	skew = 0                                 // the server time is corrected

	c := newTestClient(func(int64, int32, bin.Encoder) (bin.Encoder, error) { return nil, nil })
	c.messageID = gen
	c.serverTime = NewServerTime(clk, 0)
	state := NewSessionState()
	a.True(state.TryAcquire())
	c.sessionState = state
	key := crypto.Key{1}.WithID()
	c.authKey = key
	c.sessionID = 42

	a.Greater(gen.New(proto.MessageFromClient), high, "ids keep increasing within the session")
	var b bin.Buffer
	a.NoError(b.Encode(&mt.BadMsgNotification{BadMsgID: high, ErrorCode: codeMessageIDTooHigh}))
	a.NoError(c.handleBadMsg(int64(proto.NewMessageID(clk.Now(), proto.MessageFromServer)), &b))

	select {
	case <-c.sessionFailed.Ready():
	default:
		a.Fail("session not failed")
	}
	a.Less(gen.New(proto.MessageFromClient), high, "ids follow the corrected time again")
	c.releaseSessionState()
	_, _, ok := state.load(key.ID)
	a.False(ok, "failed session must not be continued")
}
