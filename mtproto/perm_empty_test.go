package mtproto

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/log"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/proto"
	"github.com/gotd/td/rpc"
	"github.com/gotd/td/tdsync"
	"github.com/gotd/td/tgerr"
)

func permEmptyResult(t *testing.T, reqID int64, code int, message string) *bin.Buffer {
	t.Helper()
	var body bin.Buffer
	require.NoError(t, (&mt.RPCError{ErrorCode: code, ErrorMessage: message}).Encode(&body))
	var b bin.Buffer
	require.NoError(t, (&proto.Result{RequestMessageID: reqID, Result: body.Raw()}).Encode(&b))
	return &b
}

func runPendingRequest(t *testing.T, e *rpc.Engine, id int64) <-chan error {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		done <- e.Do(context.Background(), rpc.Request{MsgID: id, Input: &mt.PingRequest{}, Output: &mt.Pong{}})
	}()
	return done
}

// AUTH_KEY_PERM_EMPTY in PFS mode: the request fails as retryable on a new
// connection and the connection asks to be recreated (new temporary key).
func TestHandleResultPermEmptyPFS(t *testing.T) {
	e := rpc.New(rpc.NopSend, rpc.Options{RetryInterval: time.Minute, Clock: clock.System})
	defer e.ForceClose()
	conn := Conn{
		pfs:            true,
		rpc:            e,
		log:            log.For(log.Nop),
		pfsBindingLost: tdsync.NewReady(),
	}
	done := runPendingRequest(t, e, 1)
	time.Sleep(10 * time.Millisecond) // Let Do register the request.
	require.NoError(t, conn.handleResult(permEmptyResult(t, 1, 401, "AUTH_KEY_PERM_EMPTY")))
	err := <-done
	require.ErrorIs(t, err, rpc.ErrEngineClosed)
	select {
	case <-conn.pfsBindingLost.Ready():
	default:
		t.Fatal("reconnect not requested")
	}
}

// Without PFS the error is passed to the caller as is.
func TestHandleResultPermEmptyNoPFS(t *testing.T) {
	e := rpc.New(rpc.NopSend, rpc.Options{RetryInterval: time.Minute, Clock: clock.System})
	defer e.ForceClose()
	conn := Conn{rpc: e, log: log.For(log.Nop), pfsBindingLost: tdsync.NewReady()}
	done := runPendingRequest(t, e, 2)
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, conn.handleResult(permEmptyResult(t, 2, 401, "AUTH_KEY_PERM_EMPTY")))
	require.True(t, tgerr.Is(<-done, "AUTH_KEY_PERM_EMPTY"))
}
