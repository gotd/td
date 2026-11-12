package mtproto

import (
	"context"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/rpc"
)

// Invoke sends input and decodes result into output.
//
// NOTE: Assuming that call contains content message (seqno increment).
func (c *Conn) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	timeResends := 0
	for {
		err := c.invokeOnce(ctx, input, output)
		if errors.Is(err, errSessionReset) {
			// The session was replaced while the request was pending: send it
			// again with a new msg_id and seqno of the new session.
			c.log.Info(ctx, "Resending request in new session")
			continue
		}
		if code, ok := badMsgTimeCode(err); ok && timeResends < maxTimeResends {
			// The server refused the msg_id because of its time; the server
			// time offset is corrected by now (see handleBadMsg and
			// observeServerTime), so send the request again with a new
			// msg_id, like TDLib does for 16, 17 and 20.
			timeResends++
			c.log.Info(ctx, "Resending request with corrected msg_id",
				log.Int("code", code),
				log.Duration("server_time_offset", c.serverTimeOffset()),
			)
			continue
		}
		return err
	}
}

// maxTimeResends limits resends of one request after bad_msg_notification
// about its msg_id time.
const maxTimeResends = 3

// codeMessageTooOld is bad_msg_notification "message too old".
const codeMessageTooOld = 20

// badMsgTimeCode reports whether err is a bad_msg_notification about the
// msg_id time (16, 17 or 20), after which the request can be sent again.
func badMsgTimeCode(err error) (int, bool) {
	var bad *badMessageError
	if !errors.As(err, &bad) {
		return 0, false
	}
	switch bad.Code {
	case codeMessageIDTooLow, codeMessageIDTooHigh, codeMessageTooOld:
		return bad.Code, true
	}
	return 0, false
}

// serverTimeOffset returns the server time offset, zero if unknown.
func (c *Conn) serverTimeOffset() time.Duration {
	if c.serverTime == nil {
		return 0
	}
	return c.serverTime.Offset()
}

func (c *Conn) invokeOnce(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	msgID, seqNo := c.nextMsgSeq(true)
	req := rpc.Request{
		MsgID:  msgID,
		SeqNo:  seqNo,
		Input:  input,
		Output: output,
	}

	logger := c.log.With(
		log.Int64("msg_id", req.MsgID),
	)
	logger.Debug(ctx, "Invoke start")
	defer logger.Debug(ctx, "Invoke end")

	if c.wire == nil {
		return c.invoke(ctx, req)
	}
	start := c.traceRequestStart(req)
	err := c.invoke(ctx, req)
	c.traceRequestDone(req, start, err)
	return err
}

func (c *Conn) invoke(ctx context.Context, req rpc.Request) error {
	if err := c.rpc.Do(ctx, req); err != nil {
		var badMsgErr *badMessageError
		if errors.As(err, &badMsgErr) && badMsgErr.Code == codeIncorrectServerSalt {
			// Should retry with new salt.
			c.log.Debug(ctx, "Setting server salt")
			// Store salt from server.
			c.storeSalt(badMsgErr.NewSalt)
			// Reset saved salts to fetch new.
			c.salts.Reset()
			c.log.Info(ctx, "Retrying request after basMsgErr", log.Int64("msg_id", req.MsgID))
			c.traceResend(req, "bad_server_salt")
			return c.rpc.Do(ctx, req)
		}
		return errors.Wrap(err, "rpcDoRequest")
	}

	return nil
}

func (c *Conn) dropRPC(req rpc.Request) error {
	ctx, cancel := context.WithTimeout(context.Background(),
		c.getTimeout(mt.RPCDropAnswerRequestTypeID),
	)
	defer cancel()

	var resp mt.RPCDropAnswerBox
	if err := c.Invoke(ctx, &mt.RPCDropAnswerRequest{
		ReqMsgID: req.MsgID,
	}, &resp); err != nil {
		return err
	}

	switch resp.RpcDropAnswer.(type) {
	case *mt.RPCAnswerDropped, *mt.RPCAnswerDroppedRunning:
		return nil
	case *mt.RPCAnswerUnknown:
		return errors.New("answer unknown")
	default:
		return errors.Errorf("unexpected response type: %T", resp.RpcDropAnswer)
	}
}
