package rpc

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		// Waiting request from client.
		assert.Equal(t, request{
			MsgID: reqID,
			SeqNo: seqNo,
			Input: &mt.PingRequest{PingID: pingID},
		}, <-incoming)

		// Make sure that client calls time.After
		// before time travel
		<-observer

		// Simulate job.
		clock.Travel(time.Second)

		// Notify client about error.
		e.NotifyError(reqID, expectedErr)

		return nil
	}

	client := func(t *testing.T, e *Engine) error {
		err := e.Do(context.TODO(), Request{
			ID:       reqID,
			Sequence: seqNo,
			Input: &mt.PingRequest{
				PingID: pingID,
			},
		})
		assert.True(t, errors.Is(err, expectedErr), "expected error")

		return nil
	}

	runTest(t, Config{
		RetryInterval: time.Second * 3,
		MaxRetries:    2,
		Clock:         clock,
	}, server, client)
}

func TestRPCResult(t *testing.T) {
	clock := neo.NewTime(defaultNow)
	observer := clock.Observe()

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		// Waiting request from client.
		assert.Equal(t, request{
			MsgID: reqID,
			SeqNo: seqNo,
			Input: &mt.PingRequest{PingID: pingID},
		}, <-incoming)

		// Make sure that engine calls time.After
		// before time travel.
		<-observer

		// Simulate job.
		clock.Travel(time.Second * 2)

		var b bin.Buffer
		if err := b.Encode(&mt.Pong{
			MsgID:  reqID,
			PingID: pingID,
		}); err != nil {
			return err
		}

		// Send response.
		return e.NotifyResult(reqID, &b)
	}

	client := func(t *testing.T, e *Engine) error {
		var out mt.Pong
		err := e.Do(context.TODO(), Request{
			ID:       reqID,
			Sequence: seqNo,
			Input:    &mt.PingRequest{PingID: pingID},
			Output:   &out,
		})

		assert.NoError(t, err)
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
	}, server, client)
}

func TestRPCAckThenResult(t *testing.T) {
	clock := neo.NewTime(defaultNow)
	observer := clock.Observe()

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		// Wait request from client.
		assert.Equal(t, request{
			MsgID: 1,
			SeqNo: 1,
			Input: &mt.PingRequest{PingID: pingID},
		}, <-incoming)

		// Make sure that client calls time.After
		// before time travel.
		<-observer

		// Simulate request processing.
		clock.Travel(time.Second * 2)

		// Acknowledge request.
		e.NotifyAcks([]int64{reqID})

		// Simulate request processing.
		clock.Travel(time.Second * 1)

		var b bin.Buffer
		if err := b.Encode(&mt.Pong{
			MsgID:  reqID,
			PingID: pingID,
		}); err != nil {
			return err
		}

		// Send response.
		return e.NotifyResult(reqID, &b)
	}

	client := func(t *testing.T, e *Engine) error {
		var out mt.Pong
		assert.NoError(t, e.Do(context.TODO(), Request{
			ID:       reqID,
			Sequence: seqNo,
			Input:    &mt.PingRequest{PingID: pingID},
			Output:   &out,
		}))
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
	}, server, client)
}

func TestRPCAckWithRetryResult(t *testing.T) {
	clock := neo.NewTime(defaultNow)
	observer := clock.Observe()

	server := func(t *testing.T, e *Engine, incoming <-chan request) error {
		// Waiting request from client.
		assert.Equal(t, request{
			MsgID: reqID,
			SeqNo: seqNo,
			Input: &mt.PingRequest{PingID: pingID},
		}, <-incoming)

		// Make sure that client calls time.After
		// before time travel.
		<-observer

		// Simulate request loss.
		//
		// Client have retry interval set to 4s,
		// so we must receive request again.
		clock.Travel(time.Second * 6)

		// Receive request.
		assert.Equal(t, request{
			MsgID: reqID,
			SeqNo: seqNo,
			Input: &mt.PingRequest{PingID: pingID},
		}, <-incoming)

		var b bin.Buffer
		if err := b.Encode(&mt.Pong{
			MsgID:  reqID,
			PingID: pingID,
		}); err != nil {
			return err
		}

		// Send response.
		return e.NotifyResult(reqID, &b)
	}

	client := func(t *testing.T, e *Engine) error {
		var out mt.Pong
		assert.NoError(t, e.Do(context.TODO(), Request{
			ID:       1,
			Sequence: 1,
			Input:    &mt.PingRequest{PingID: pingID},
			Output:   &out,
		}))
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
		requests <- request{
			MsgID: msgID,
			SeqNo: seqNo,
			Input: in,
		}
		return nil
	}, cfg)

	var g errgroup.Group
	g.Go(func() error { return server(t, e, requests) })
	g.Go(func() error { return client(t, e) })

	assert.NoError(t, g.Wait())
}
