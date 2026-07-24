package rpc

import (
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/log/logzap"
	"github.com/gotd/neo"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/mt"
	"github.com/gotd/td/testutil"
	"github.com/gotd/td/transport"
)

type request struct {
	MsgID int64
	SeqNo int32
	Input bin.Encoder
}

var defaultNow = testutil.Date()

const (
	msgID  int64 = 1
	pingID int64 = 1337
	seqNo  int32 = 1
)

func TestRPCError(t *testing.T) {
	clock := neo.NewTime(defaultNow)
	observer := clock.Observe()
	expectedErr := errors.New("server side error")
	log := zaptest.NewLogger(t)

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		log := log.Named("server")

		log.Info("Waiting ping request")
		require.Equal(t, request{
			MsgID: msgID,
			SeqNo: seqNo,
			Input: &mt.PingRequest{PingID: pingID},
		}, <-incoming)

		log.Info("Got ping request")

		// Make sure that client calls time.After
		// before time travel
		<-observer

		log.Info("Traveling into the future for a second (simulate job)")
		clock.Travel(time.Second)

		log.Info("Sending RPC error")
		e.NotifyError(msgID, expectedErr)

		return nil
	}

	client := func(t *testing.T, e *Engine) error {
		log := log.Named("client")

		log.Info("Sending ping request")
		err := e.Do(context.TODO(), Request{
			MsgID: msgID,
			SeqNo: seqNo,
			Input: &mt.PingRequest{
				PingID: pingID,
			},
		})

		log.Info("Got pong response")
		require.True(t, errors.Is(err, expectedErr), "expected error")

		return nil
	}

	runTest(t, Options{
		RetryInterval: time.Second * 3,
		MaxRetries:    2,
		Clock:         clock,
		Logger:        logzap.New(log.Named("rpc")),
	}, server, client)
}

func TestRPCResult(t *testing.T) {
	clock := neo.NewTime(defaultNow)
	observer := clock.Observe()
	log := zaptest.NewLogger(t)

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		log := log.Named("server")

		log.Info("Waiting ping request")
		require.Equal(t, request{
			MsgID: msgID,
			SeqNo: seqNo,
			Input: &mt.PingRequest{PingID: pingID},
		}, <-incoming)

		log.Info("Got ping request")
		// Make sure that engine calls time.After
		// before time travel.
		<-observer

		log.Info("Traveling into the future for 2 seconds (simulate job)")
		clock.Travel(time.Second * 2)

		var b bin.Buffer
		if err := b.Encode(&mt.Pong{
			MsgID:  msgID,
			PingID: pingID,
		}); err != nil {
			return err
		}

		log.Info("Sending pong response")
		return e.NotifyResult(msgID, &b)
	}

	client := func(t *testing.T, e *Engine) error {
		log := log.Named("client")

		log.Info("Sending ping request")
		var out mt.Pong
		require.NoError(t, e.Do(context.TODO(), Request{
			MsgID:  msgID,
			SeqNo:  seqNo,
			Input:  &mt.PingRequest{PingID: pingID},
			Output: &out,
		}))

		log.Info("Got pong response")
		require.Equal(t, mt.Pong{
			MsgID:  msgID,
			PingID: pingID,
		}, out)

		return nil
	}

	runTest(t, Options{
		RetryInterval: time.Second * 4,
		MaxRetries:    2,
		Clock:         clock,
		Logger:        logzap.New(log.Named("rpc")),
	}, server, client)
}

func TestRPCAckThenResult(t *testing.T) {
	clock := neo.NewTime(defaultNow)
	observer := clock.Observe()
	log := zaptest.NewLogger(t)

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		log := log.Named("server")

		log.Info("Waiting ping request")
		require.Equal(t, request{
			MsgID: msgID,
			SeqNo: seqNo,
			Input: &mt.PingRequest{PingID: pingID},
		}, <-incoming)

		// Make sure that client calls time.After
		// before time travel.
		<-observer

		log.Info("Traveling into the future for 2 seconds (simulate job)")
		clock.Travel(time.Second * 2)

		log.Info("Sending ACK")
		e.NotifyAcks([]int64{msgID})

		log.Info("Traveling into the future for 6 seconds (simulate request processing)")
		clock.Travel(time.Second * 6)

		var b bin.Buffer
		if err := b.Encode(&mt.Pong{
			MsgID:  msgID,
			PingID: pingID,
		}); err != nil {
			return err
		}

		log.Info("Sending response")
		return e.NotifyResult(msgID, &b)
	}

	client := func(t *testing.T, e *Engine) error {
		log := log.Named("client")

		log.Info("Sending ping request")
		var out mt.Pong
		require.NoError(t, e.Do(context.TODO(), Request{
			MsgID:  msgID,
			SeqNo:  seqNo,
			Input:  &mt.PingRequest{PingID: pingID},
			Output: &out,
		}))

		log.Info("Got pong response")
		require.Equal(t, mt.Pong{
			MsgID:  msgID,
			PingID: pingID,
		}, out)

		return nil
	}

	runTest(t, Options{
		RetryInterval: time.Second * 4,
		MaxRetries:    2,
		Clock:         clock,
		Logger:        logzap.New(log.Named("rpc")),
	}, server, client)
}

func TestRPCWithRetryResult(t *testing.T) {
	clock := neo.NewTime(defaultNow)
	observer := clock.Observe()
	log := zaptest.NewLogger(t)

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		log := log.Named("server")

		log.Info("Waiting ping request")
		require.Equal(t, request{
			MsgID: msgID,
			SeqNo: seqNo,
			Input: &mt.PingRequest{PingID: pingID},
		}, <-incoming)
		log.Info("Got ping request")

		// Make sure that client calls time.After
		// before time travel.
		<-observer

		log.Info("Traveling into the future for 6 seconds (simulate request loss)")
		clock.Travel(time.Second * 6)

		log.Info("Waiting re-sending request")
		require.Equal(t, request{
			MsgID: msgID,
			SeqNo: seqNo,
			Input: &mt.PingRequest{PingID: pingID},
		}, <-incoming)
		log.Info("Got ping request")

		var b bin.Buffer
		if err := b.Encode(&mt.Pong{
			MsgID:  msgID,
			PingID: pingID,
		}); err != nil {
			return err
		}

		log.Info("Send pong response")
		return e.NotifyResult(msgID, &b)
	}

	client := func(t *testing.T, e *Engine) error {
		log := log.Named("client")

		log.Info("Sending ping request")
		var out mt.Pong
		require.NoError(t, e.Do(context.TODO(), Request{
			MsgID:  1,
			SeqNo:  seqNo,
			Input:  &mt.PingRequest{PingID: pingID},
			Output: &out,
		}))

		log.Info("Got pong response")
		require.Equal(t, mt.Pong{
			MsgID:  msgID,
			PingID: pingID,
		}, out)

		return nil
	}

	runTest(t, Options{
		RetryInterval: time.Second * 4,
		MaxRetries:    5,
		Clock:         clock,
		Logger:        logzap.New(log.Named("rpc")),
	}, server, client)
}

func TestEngineGracefulShutdown(t *testing.T) {
	var (
		log             = zaptest.NewLogger(t)
		expectedErr     = errors.New("server side error")
		requestsCount   = 10
		serverRecv      sync.WaitGroup
		canSendResponse sync.Mutex
	)

	serverRecv.Add(requestsCount)
	canSendResponse.Lock()

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		log := log.Named("server")

		var batch []request
		for i := 0; i < requestsCount; i++ {
			batch = append(batch, <-incoming)
			serverRecv.Done()
		}
		e.log.Info(context.Background(), "Got all requests")

		canSendResponse.Lock()
		e.log.Info(context.Background(), "Sending responses")
		for _, req := range batch {
			log.Info("send response")
			e.NotifyError(req.MsgID, expectedErr)
		}
		canSendResponse.Unlock()

		return nil
	}

	client := func(t *testing.T, e *Engine) error {
		var currMsgID int64

		for i := 0; i < requestsCount; i++ {
			go func(t *testing.T, msgID int64) {
				var out mt.Pong
				require.Equal(t, e.Do(context.TODO(), Request{
					MsgID:  msgID,
					SeqNo:  seqNo,
					Input:  &mt.PingRequest{PingID: pingID},
					Output: &out,
				}), expectedErr)
			}(t, currMsgID)

			currMsgID++
		}

		// wait until server receive all requests
		serverRecv.Wait()
		// allow server to send responses
		canSendResponse.Unlock()
		// close the engine
		e.Close()

		return nil
	}

	runTest(t, Options{
		RetryInterval: time.Second * 5,
		MaxRetries:    5,
		Logger:        logzap.New(log.Named("rpc")),
	}, server, client)
}

func TestDropRPC(t *testing.T) {
	clock := neo.NewTime(defaultNow)
	log := zaptest.NewLogger(t)
	serverRecvRequest := make(chan struct{})
	clientCancelledCtx := make(chan struct{})
	dropChan := make(chan Request)

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		log := log.Named("server")

		log.Info("Waiting ping request")
		require.Equal(t, request{
			MsgID: msgID,
			SeqNo: seqNo,
			Input: &mt.PingRequest{PingID: pingID},
		}, <-incoming)

		close(serverRecvRequest)
		<-clientCancelledCtx

		log.Info("Waiting drop request")
		require.Equal(t, msgID, (<-dropChan).MsgID)
		return nil
	}

	client := func(t *testing.T, e *Engine) error {
		log := log.Named("client")

		log.Info("Sending ping request")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			<-serverRecvRequest
			log.Info("Canceling request context")
			cancel()
			close(clientCancelledCtx)
		}()

		require.ErrorIs(t, e.Do(ctx, Request{
			MsgID:  msgID,
			SeqNo:  seqNo,
			Input:  &mt.PingRequest{PingID: pingID},
			Output: &mt.Pong{},
		}), context.Canceled)

		return nil
	}

	runTest(t, Options{
		RetryInterval: time.Second * 4,
		MaxRetries:    2,
		Clock:         clock,
		Logger:        logzap.New(log.Named("rpc")),
		DropHandler:   func(req Request) error { dropChan <- req; return nil },
	}, server, client)
}

func TestEngineForceCloseNotAcked(t *testing.T) {
	log := zaptest.NewLogger(t)
	gotRequest := make(chan struct{})
	result := make(chan error, 1)

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		// Server receives the request, but never acknowledges it.
		<-incoming
		close(gotRequest)
		return nil
	}

	client := func(t *testing.T, e *Engine) error {
		go func() {
			result <- e.Do(context.TODO(), Request{
				MsgID:  msgID,
				SeqNo:  seqNo,
				Input:  &mt.PingRequest{PingID: pingID},
				Output: &mt.Pong{},
			})
		}()
		<-gotRequest
		e.ForceClose()

		// Sent but not acknowledged request is safe to resend, so the error
		// must be retryable (ErrEngineClosed) and not a plain cancellation.
		err := <-result
		require.ErrorIs(t, err, ErrEngineClosed)
		require.NotErrorIs(t, err, context.Canceled)
		return nil
	}

	runTest(t, Options{
		RetryInterval: time.Minute,
		MaxRetries:    2,
		Logger:        logzap.New(log.Named("rpc")),
	}, server, client)
}

func TestEngineForceCloseAcked(t *testing.T) {
	log := zaptest.NewLogger(t)
	acked := make(chan struct{})
	result := make(chan error, 1)

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		// Server acknowledges the request, but never responds.
		req := <-incoming
		e.NotifyAcks([]int64{req.MsgID})
		close(acked)
		return nil
	}

	client := func(t *testing.T, e *Engine) error {
		go func() {
			result <- e.Do(context.TODO(), Request{
				MsgID:  msgID,
				SeqNo:  seqNo,
				Input:  &mt.PingRequest{PingID: pingID},
				Output: &mt.Pong{},
			})
		}()
		<-acked
		e.ForceClose()

		// Acknowledged request may be already processed by server, so
		// transparent resend is not safe: plain cancellation is expected.
		err := <-result
		require.ErrorIs(t, err, context.Canceled)
		require.NotErrorIs(t, err, ErrEngineClosed)
		return nil
	}

	runTest(t, Options{
		RetryInterval: time.Minute,
		MaxRetries:    2,
		Logger:        logzap.New(log.Named("rpc")),
	}, server, client)
}

// TestEngineForceCloseAckedSentinel pins the diagnosability property added
// on top of TestEngineForceCloseAcked: the acked-but-unanswered error must
// simultaneously (1) still satisfy errors.Is(err, context.Canceled), so
// existing callers checking for plain cancellation keep working; (2) still
// NOT satisfy errors.Is(err, ErrEngineClosed), so it is not misclassified as
// retryable; and (3) satisfy errors.Is(err, ErrEngineClosedAfterAck), so
// callers can distinguish "closed after my request was acknowledged" from a
// caller-initiated cancellation.
func TestEngineForceCloseAckedSentinel(t *testing.T) {
	log := zaptest.NewLogger(t)
	acked := make(chan struct{})
	result := make(chan error, 1)

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		// Server acknowledges the request, but never responds.
		req := <-incoming
		e.NotifyAcks([]int64{req.MsgID})
		close(acked)
		return nil
	}

	client := func(t *testing.T, e *Engine) error {
		go func() {
			result <- e.Do(context.TODO(), Request{
				MsgID:  msgID,
				SeqNo:  seqNo,
				Input:  &mt.PingRequest{PingID: pingID},
				Output: &mt.Pong{},
			})
		}()
		<-acked
		e.ForceClose()

		err := <-result
		require.ErrorIs(t, err, context.Canceled)
		require.NotErrorIs(t, err, ErrEngineClosed)
		require.ErrorIs(t, err, ErrEngineClosedAfterAck)
		// The rendered message must name the acknowledgement, so an operator
		// reading it for a stuck request does not mistake it for a plain
		// caller-initiated cancellation.
		require.Contains(t, err.Error(), "engine forcibly closed after request was acknowledged: ")
		return nil
	}

	runTest(t, Options{
		RetryInterval: time.Minute,
		MaxRetries:    2,
		Logger:        logzap.New(log.Named("rpc")),
	}, server, client)
}

// retryAckRaceEngine builds an Engine whose send succeeds once and then fails
// on every retry, modelling the branch's own target scenario: the TCP send
// buffer is full while the read path is still live. onRetry runs inside the
// failing retry send, before the failure is returned, so a test can deliver an
// acknowledge that races the write failure.
func retryAckRaceEngine(t *testing.T, clk clock.Clock, onRetry func(e *Engine)) (*Engine, <-chan struct{}) {
	t.Helper()

	var (
		e     *Engine
		sends atomic.Int64
		first = make(chan struct{})
	)

	e = New(func(ctx context.Context, msgID int64, seqNo int32, in bin.Encoder) error {
		if sends.Add(1) == 1 {
			// First send reaches the server.
			close(first)
			return nil
		}
		if onRetry != nil {
			onRetry(e)
		}
		// The retry parks on the wedged socket and eventually fails. This is
		// modelled on the error shape transport.Send produces; it is not
		// identical. transport.Send uses a hand-rolled single-line error type
		// with a single-valued Unwrap, whereas errors.Join renders across two
		// lines and unwraps to a slice. Both are masked identically by
		// retrySendError, which is what this test exercises — do not read this
		// literal as documentation of how transport renders write errors.
		return errors.Wrap(errors.Join(transport.ErrWriteFailed, os.ErrDeadlineExceeded), "write")
	}, Options{
		RetryInterval: time.Second * 4,
		MaxRetries:    5,
		Clock:         clk,
		Logger:        logzap.New(zaptest.NewLogger(t).Named("rpc")),
	})

	return e, first
}

// TestRetrySendFailedAfterAck pins the fix for the duplicate-RPC hazard: send
// #1 reaches the server, no ack arrives within RetryInterval, and the retry
// send delivers the ack over the still-live read path before failing with
// transport.ErrWriteFailed.
//
// The request IS acknowledged, so the write failure must be discarded
// entirely. Reporting it would surface an ErrWriteFailed-matching error, which
// telegram/invoke.go and pool/pool_conn.go classify as retryable and would
// resend under a fresh msg_id — duplicating an RPC the server already holds.
func TestRetrySendFailedAfterAck(t *testing.T) {
	clk := neo.NewTime(defaultNow)
	observer := clk.Observe()

	e, first := retryAckRaceEngine(t, clk, func(e *Engine) {
		// The ack for send #1 lands while the retry write is still parked.
		e.NotifyAcks([]int64{msgID})

		// The server then answers the (single, acknowledged) request.
		var b bin.Buffer
		require.NoError(t, b.Encode(&mt.Pong{MsgID: msgID, PingID: pingID}))
		require.NoError(t, e.NotifyResult(msgID, &b))
	})
	defer e.Close()

	result := make(chan error, 1)
	var out mt.Pong
	go func() {
		result <- e.Do(context.TODO(), Request{
			MsgID:  msgID,
			SeqNo:  seqNo,
			Input:  &mt.PingRequest{PingID: pingID},
			Output: &out,
		})
	}()

	<-first
	// Wait until Do is parked on the retry timer, then fire it.
	<-observer
	clk.Travel(time.Second * 6)

	err := <-result
	require.NoError(t, err)
	require.NotErrorIs(t, err, transport.ErrWriteFailed)
	require.Equal(t, mt.Pong{MsgID: msgID, PingID: pingID}, out)
}

// TestRetrySendFailedWithoutAck covers the sibling case: the retry send fails
// and no acknowledge ever arrives. Send #1 still reached the wire, so the
// server may hold the request and a transparent resend under a fresh msg_id
// would bypass MTProto deduplication. The error must therefore not be
// classified as retryable, while staying identifiable via ErrRetrySendFailed
// and still naming the underlying write failure in its message.
func TestRetrySendFailedWithoutAck(t *testing.T) {
	clk := neo.NewTime(defaultNow)
	observer := clk.Observe()

	e, first := retryAckRaceEngine(t, clk, nil)
	defer e.Close()

	result := make(chan error, 1)
	go func() {
		result <- e.Do(context.TODO(), Request{
			MsgID:  msgID,
			SeqNo:  seqNo,
			Input:  &mt.PingRequest{PingID: pingID},
			Output: &mt.Pong{},
		})
	}()

	<-first
	<-observer
	clk.Travel(time.Second * 6)

	err := <-result
	require.Error(t, err)
	require.NotErrorIs(t, err, transport.ErrWriteFailed)
	require.ErrorIs(t, err, ErrRetrySendFailed)
	// The cause must survive for debugging even though the chain is cut.
	require.Contains(t, err.Error(), transport.ErrWriteFailed.Error())
	// Joined with ": ", not with the newline errors.Join would insert.
	require.Contains(t, err.Error(), ErrRetrySendFailed.Error()+": ")
}

func runTest(
	t *testing.T,
	cfg Options,
	server func(t *testing.T, e *Engine, incoming <-chan request) error,
	client func(t *testing.T, e *Engine) error,
) {
	t.Helper()

	// Channel of client requests sent to the server.
	requests := make(chan request)
	defer close(requests)

	e := New(func(ctx context.Context, msgID int64, seqNo int32, in bin.Encoder) error {
		req := request{
			MsgID: msgID,
			SeqNo: seqNo,
			Input: in,
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case requests <- req:
			return nil
		}
	}, cfg)

	var g errgroup.Group
	g.Go(func() error { return server(t, e, requests) })
	g.Go(func() error { return client(t, e) })

	require.NoError(t, g.Wait())
	e.Close()
}
