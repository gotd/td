package rpc

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/neo"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/mt"
)

type request struct {
	MsgID int64
	SeqNo int32
	Input bin.Encoder
}

var defaultNow = time.Date(2010, 10, 10, 3, 45, 12, 23, time.UTC)

const (
	reqID  = 1
	pingID = 1337
	seqNo  = 1
)

func TestRPCError(t *testing.T) {
	clock := neo.NewTime(defaultNow)
	observer := clock.Observe()
	expectedErr := errors.New("server side error")
	log := zaptest.NewLogger(t)

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		log := log.Named("server")

		log.Info("Waiting ping request")
		assert.Equal(t, request{
			MsgID: reqID,
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
		e.NotifyError(reqID, expectedErr)

		return nil
	}

	client := func(t *testing.T, e *Engine) error {
		log := log.Named("client")

		log.Info("Sending ping request")
		err := e.Do(context.TODO(), Request{
			ID:       reqID,
			Sequence: seqNo,
			Input: &mt.PingRequest{
				PingID: pingID,
			},
		})

		log.Info("Got pong response")
		assert.True(t, errors.Is(err, expectedErr), "expected error")

		return nil
	}

	runTest(t, Config{
		RetryInterval: time.Second * 3,
		MaxRetries:    2,
		Clock:         clock,
		Logger:        log.Named("rpc"),
	}, server, client)
}

func TestRPCResult(t *testing.T) {
	clock := neo.NewTime(defaultNow)
	observer := clock.Observe()
	log := zaptest.NewLogger(t)

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		log := log.Named("server")

		log.Info("Waiting ping request")
		assert.Equal(t, request{
			MsgID: reqID,
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
			MsgID:  reqID,
			PingID: pingID,
		}); err != nil {
			return err
		}

		log.Info("Sending pong response")
		return e.NotifyResult(reqID, &b)
	}

	client := func(t *testing.T, e *Engine) error {
		log := log.Named("client")

		log.Info("Sending ping request")
		var out mt.Pong
		assert.NoError(t, e.Do(context.TODO(), Request{
			ID:       reqID,
			Sequence: seqNo,
			Input:    &mt.PingRequest{PingID: pingID},
			Output:   &out,
		}))

		log.Info("Got pong response")
		assert.Equal(t, mt.Pong{
			MsgID:  reqID,
			PingID: pingID,
		}, out)

		return nil
	}

	runTest(t, Config{
		RetryInterval: time.Second * 4,
		MaxRetries:    2,
		Clock:         clock,
		Logger:        log.Named("rpc"),
	}, server, client)
}

func TestRPCAckThenResult(t *testing.T) {
	clock := neo.NewTime(defaultNow)
	observer := clock.Observe()
	log := zaptest.NewLogger(t)

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		log := log.Named("server")

		log.Info("Waiting ping request")
		assert.Equal(t, request{
			MsgID: reqID,
			SeqNo: seqNo,
			Input: &mt.PingRequest{PingID: pingID},
		}, <-incoming)

		// Make sure that client calls time.After
		// before time travel.
		<-observer

		log.Info("Traveling into the future for 2 seconds (simulate job)")
		clock.Travel(time.Second * 2)

		log.Info("Sending ACK")
		e.NotifyAcks([]int64{reqID})

		log.Info("Traveling into the future for 6 seconds (simulate request processing)")
		clock.Travel(time.Second * 6)

		var b bin.Buffer
		if err := b.Encode(&mt.Pong{
			MsgID:  reqID,
			PingID: pingID,
		}); err != nil {
			return err
		}

		log.Info("Sending response")
		return e.NotifyResult(reqID, &b)
	}

	client := func(t *testing.T, e *Engine) error {
		log := log.Named("client")

		log.Info("Sending ping request")
		var out mt.Pong
		assert.NoError(t, e.Do(context.TODO(), Request{
			ID:       reqID,
			Sequence: seqNo,
			Input:    &mt.PingRequest{PingID: pingID},
			Output:   &out,
		}))

		log.Info("Got pong response")
		assert.Equal(t, mt.Pong{
			MsgID:  reqID,
			PingID: pingID,
		}, out)

		return nil
	}

	runTest(t, Config{
		RetryInterval: time.Second * 4,
		MaxRetries:    2,
		Clock:         clock,
		Logger:        log.Named("rpc"),
	}, server, client)
}

func TestRPCAckWithRetryResult(t *testing.T) {
	clock := neo.NewTime(defaultNow)
	observer := clock.Observe()
	log := zaptest.NewLogger(t)

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		log := log.Named("server")

		log.Info("Waiting ping request")
		assert.Equal(t, request{
			MsgID: reqID,
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
		assert.Equal(t, request{
			MsgID: reqID,
			SeqNo: seqNo,
			Input: &mt.PingRequest{PingID: pingID},
		}, <-incoming)
		log.Info("Got ping request")

		var b bin.Buffer
		if err := b.Encode(&mt.Pong{
			MsgID:  reqID,
			PingID: pingID,
		}); err != nil {
			return err
		}

		log.Info("Send pong response")
		return e.NotifyResult(reqID, &b)
	}

	client := func(t *testing.T, e *Engine) error {
		log := log.Named("client")

		log.Info("Sending ping request")
		var out mt.Pong
		assert.NoError(t, e.Do(context.TODO(), Request{
			ID:       1,
			Sequence: 1,
			Input:    &mt.PingRequest{PingID: pingID},
			Output:   &out,
		}))

		log.Info("Got pong response")
		assert.Equal(t, mt.Pong{
			MsgID:  reqID,
			PingID: pingID,
		}, out)

		return nil
	}

	runTest(t, Config{
		RetryInterval: time.Second * 4,
		MaxRetries:    5,
		Clock:         clock,
		Logger:        log.Named("rpc"),
	}, server, client)
}

func runTest(
	t *testing.T,
	cfg Config,
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

	assert.NoError(t, g.Wait())
	e.Close()
	assert.NoError(t, cfg.Logger.Sync())
}
