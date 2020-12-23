package rpc

import (
	"fmt"
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

func sendTo(c chan<- request) Send {
	return func(ctx context.Context, msgID int64, seqNo int32, in bin.Encoder) error {
		c <- request{
			MsgID: msgID,
			SeqNo: seqNo,
			Input: in,
		}
		return nil
	}
}

var defaultNow = time.Date(2010, 10, 10, 3, 45, 12, 23, time.UTC)

func TestRPCError(t *testing.T) {
	var (
		// What server expect.
		serverExpect = request{
			MsgID: 1,
			SeqNo: 1,
			Input: &mt.PingRequest{PingID: 1337},
		}
		// What server should return.
		serverErrorResponse = fmt.Errorf("omg")
	)

	// Channel of client requests sent to the server.
	requests := make(chan request)
	defer close(requests)

	clock := neo.NewTime(defaultNow)

	e := New(sendTo(requests), Config{
		RetryInterval: time.Second * 3,
		MaxRetries:    2,
		Clock:         clock,
	})

	g, gCtx := errgroup.WithContext(context.Background())
	observe := clock.Observe()

	g.Go(func() error {
		// Waiting request from client.
		req := <-requests
		<-observe

		// Verify client request.
		assert.Equal(t, serverExpect, req)

		// Simulate job.
		clock.Travel(time.Second)

		// NotifyAcks client about error.
		e.NotifyError(req.MsgID, serverErrorResponse)

		return nil
	})

	// Client behavior.
	g.Go(func() error {
		return e.Do(gCtx, Request{
			ID:       1,
			Sequence: 1,
			Input: &mt.PingRequest{
				PingID: 1337,
			},
		})
	})

	assert.Equal(t, serverErrorResponse, g.Wait())
}

func TestRPCResult(t *testing.T) {
	var (
		// What server expect.
		serverExpect = request{
			MsgID: 1,
			SeqNo: 1,
			Input: &mt.PingRequest{PingID: 1337},
		}
		// What server should return.
		serverResponse = mt.Pong{
			MsgID:  1,
			PingID: 1337,
		}
	)

	// Channel of client requests sent to the server.
	requests := make(chan request)
	defer close(requests)

	clock := neo.NewTime(defaultNow)

	e := New(sendTo(requests), Config{
		RetryInterval: time.Second * 3,
		MaxRetries:    2,
		Clock:         clock,
	})

	g, gCtx := errgroup.WithContext(context.Background())
	observe := clock.Observe()

	g.Go(func() error {
		// Waiting request from client.
		req := <-requests
		assert.Equal(t, serverExpect, req)

		// Simulate job.
		<-observe
		clock.Travel(time.Second * 2)

		var b bin.Buffer
		if err := b.Encode(&serverResponse); err != nil {
			return err
		}

		// Send response.
		if err := e.NotifyResult(req.MsgID, &b); err != nil {
			return err
		}

		return nil
	})

	var out mt.Pong
	// Client behavior.
	g.Go(func() error {
		return e.Do(gCtx, Request{
			ID:       1,
			Sequence: 1,
			Input:    &mt.PingRequest{PingID: 1337},
			Output:   &out,
		})
	})

	assert.NoError(t, g.Wait())
	assert.Equal(t, serverResponse, out)
}

func TestRPCAckThenResult(t *testing.T) {
	var (
		// What server expect.
		serverExpect = request{
			MsgID: 1,
			SeqNo: 1,
			Input: &mt.PingRequest{PingID: 1337},
		}
		// What server should return.
		serverResponse = mt.Pong{
			MsgID:  1,
			PingID: 1337,
		}
	)

	// Channel of client requests sent to the server.
	requests := make(chan request)
	defer close(requests)

	clock := neo.NewTime(defaultNow)

	e := New(sendTo(requests), Config{
		RetryInterval: time.Second * 4,
		MaxRetries:    2,
		Clock:         clock,
	})

	g, gCtx := errgroup.WithContext(context.Background())
	observe := clock.Observe()

	// Server behavior.
	g.Go(func() error {
		// Wait request from client.
		req := <-requests
		assert.Equal(t, serverExpect, req)

		// Simulate request processing.
		<-observe
		clock.Travel(time.Second * 2)

		// Acknowledge request.
		e.NotifyAcks([]int64{req.MsgID})

		// Simulate request processing.
		clock.Travel(time.Second * 1)

		var b bin.Buffer
		if err := b.Encode(&serverResponse); err != nil {
			return err
		}

		// Send response.
		return e.NotifyResult(req.MsgID, &b)
	})

	var out mt.Pong
	g.Go(func() error {
		return e.Do(gCtx, Request{
			ID:       1,
			Sequence: 1,
			Input:    &mt.PingRequest{PingID: 1337},
			Output:   &out,
		})
	})

	assert.NoError(t, g.Wait())
	assert.Equal(t, serverResponse, out)
}

func TestRPCAckWithRetryResult(t *testing.T) {
	var (
		// What server expect.
		serverExpect = request{
			MsgID: 1,
			SeqNo: 1,
			Input: &mt.PingRequest{PingID: 1337},
		}
		// What server should return.
		serverResponse = mt.Pong{
			MsgID:  1,
			PingID: 1337,
		}
	)

	// Channel of client requests sent to the server.
	requests := make(chan request)
	defer close(requests)

	clock := neo.NewTime(defaultNow)

	e := New(sendTo(requests), Config{
		RetryInterval: time.Second * 4,
		MaxRetries:    5,
		Clock:         clock,
	})

	g, gCtx := errgroup.WithContext(context.Background())

	// Server behavior.
	g.Go(func() error {
		observe := clock.Observe()

		// Receive request.
		req := <-requests
		assert.Equal(t, serverExpect, req)

		// Receive retry.
		<-observe
		clock.Travel(time.Second * 6)
		req = <-requests
		assert.Equal(t, serverExpect, req)

		var b bin.Buffer
		if err := b.Encode(&serverResponse); err != nil {
			return err
		}

		// Send response.
		return e.NotifyResult(req.MsgID, &b)
	})

	var out mt.Pong
	// Client behavior.
	g.Go(func() error {
		return e.Do(gCtx, Request{
			ID:       1,
			Sequence: 1,
			Input:    &mt.PingRequest{PingID: 1337},
			Output:   &out,
		})
	})

	assert.NoError(t, g.Wait())
	assert.Equal(t, serverResponse, out)
}
