package rpc

import (
	"context"

	"github.com/gotd/td/bin"
)

// Send is a function that sends requests to the server.
type Send func(ctx context.Context, msgID int64, seqNo int32, in bin.Encoder) error

// NopSend does nothing.
func NopSend(context.Context, int64, int32, bin.Encoder) error { return nil }

var _ Send = NopSend

// DropHandler handles drop rpc requests.
type DropHandler func(req Request) error

// NopDrop does nothing.
func NopDrop(Request) error { return nil }

var _ DropHandler = NopDrop
