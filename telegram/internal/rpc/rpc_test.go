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
	requests := make(chan request, 1)
	defer close(requests)

	clock := neo.NewTime(defaultNow)

	e := New(sendTo(requests), Config{
		RetryInterval: time.Second * 3,
		MaxRetries:    2,
		Clock:         clock,
	})

	var g errgroup.Group

	// Server behavior.
	g.Go(func() error {
		// Waiting request from client.
		req := <-requests

		// Verify client request.
		assert.Equal(t, serverExpect, req)

		// Simulate job.
		clock.Travel(time.Second)

		// NotifyAcks client about error.
		e.NotifyError(1, serverErrorResponse)

		return nil
	})

	// Client behavior.
	g.Go(func() error {
		return e.Do(context.Background(), Request{
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
	requests := make(chan request, 1)
	defer close(requests)

	clock := neo.NewTime(defaultNow)

	e := New(sendTo(requests), Config{
		RetryInterval: time.Second * 3,
		MaxRetries:    2,
	})

	var g errgroup.Group

	// Server behavior.
	g.Go(func() error {
		// Waiting request from client.
		req := <-requests

		// Validating request.
		assert.Equal(t, serverExpect, req)

		// Simulate job.
		clock.Travel(time.Second * 2)

		b := new(bin.Buffer)
		var serverPong mt.Pong
		serverPong.MsgID = 1
		serverPong.PingID = 1337
		if err := serverPong.Encode(b); err != nil {
			return err
		}

		// Send response.
		if err := e.NotifyResult(1, b); err != nil {
			return err
		}

		return nil
	})

	var out mt.Pong
	// Client behavior.
	g.Go(func() error {
		return e.Do(context.Background(), Request{
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
	requests := make(chan request, 1)
	defer close(requests)

	clock := neo.NewTime(defaultNow)

	e := New(sendTo(requests), Config{
		RetryInterval: time.Second * 4,
		MaxRetries:    2,
		Clock:         clock,
	})

	var g errgroup.Group

	// Server behavior.
	g.Go(func() error {
		// Wait request from client.
		req := <-requests

		// Validate request.
		assert.Equal(t, serverExpect, req)

		// Simulate job.
		clock.Travel(time.Second * 2)

		// NotifyAcks client ACK.
		e.NotifyAcks([]int64{1})

		// Simulate job again.
		clock.Travel(time.Second * 4)

		b := new(bin.Buffer)
		var serverPong mt.Pong
		serverPong.MsgID = 1
		serverPong.PingID = 1337
		if err := serverPong.Encode(b); err != nil {
			return err
		}

		// Send response.
		if err := e.NotifyResult(1, b); err != nil {
			return err
		}

		return nil
	})

	var out mt.Pong
	g.Go(func() error {
		return e.Do(context.Background(), Request{
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

	t.Skip("TODO(ernado): fix race")

	// Channel of client requests sent to the server.
	requests := make(chan request, 1)
	defer close(requests)

	clock := neo.NewTime(defaultNow)

	e := New(sendTo(requests), Config{
		RetryInterval: time.Second * 4,
		MaxRetries:    5,
		Clock:         clock,
	})

	var eg errgroup.Group

	// Server behavior.
	eg.Go(func() error {
		// Wait request from client.
		req := <-requests

		// Validate request.
		assert.Equal(t, serverExpect, req)

		// Simulate request loss.
		clock.Travel(time.Second * 6)

		// Wait that request again.
		req = <-requests

		// Validate it.
		assert.Equal(t, serverExpect, req)

		var pong mt.Pong
		pong.MsgID = 1
		pong.PingID = 1337
		b := new(bin.Buffer)
		if err := pong.Encode(b); err != nil {
			return err
		}

		// Send response.
		if err := e.NotifyResult(1, b); err != nil {
			return err
		}
		return nil
	})

	var out mt.Pong
	// Client behavior.
	eg.Go(func() error {
		return e.Do(context.Background(), Request{
			ID:       1,
			Sequence: 1,
			Input:    &mt.PingRequest{PingID: 1337},
			Output:   &out,
		})
	})

	assert.NoError(t, eg.Wait())
	assert.Equal(t, serverResponse, out)
}
