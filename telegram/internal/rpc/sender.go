package rpc

import (
	"context"

	"github.com/gotd/td/bin"
)

// Sender is an object that can send requests to the server.
type Sender interface {
	Send(ctx context.Context, msgID int64, seqNo int32, in bin.Encoder) error
}

// SendFunc is a function that sends requests to the server.
type SendFunc func(ctx context.Context, msgID int64, seqNo int32, in bin.Encoder) error

// Send sends request to the server.
func (f SendFunc) Send(ctx context.Context, msgID int64, seqNo int32, in bin.Encoder) error {
	return f(ctx, msgID, seqNo, in)
}

// NoOpSender just does nothing, always return nil
func NoOpSender() SendFunc {
	return func(ctx context.Context, msgID int64, seqNo int32, in bin.Encoder) error {
		return nil
	}
}
