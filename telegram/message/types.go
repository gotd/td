package message

import (
	"context"

	"github.com/nnqq/td/tg"
)

type multiMediaBuilder struct {
	sender *Sender
	peer   tg.InputPeerClass
	// Attached media.
	media []tg.InputSingleMedia
}

// MediaOption is an option for sending media attachments.
type MediaOption interface {
	apply(ctx context.Context, b *multiMediaBuilder) error
}

var _ MediaOption = mediaOptionFunc(nil)

// mediaOptionFunc is a function adapter for MediaOption.
type mediaOptionFunc func(ctx context.Context, b *multiMediaBuilder) error

// apply implements MediaOption.
func (m mediaOptionFunc) apply(ctx context.Context, b *multiMediaBuilder) error {
	return m(ctx, b)
}

// MultiMediaOption is an option for sending albums.
type MultiMediaOption interface {
	MediaOption
	applyMulti(ctx context.Context, b *multiMediaBuilder) error
}

type multiMediaWrapper struct {
	MediaOption
}

// applyMulti implements MultiMediaOption.
func (m multiMediaWrapper) applyMulti(ctx context.Context, b *multiMediaBuilder) error {
	return m.apply(ctx, b)
}

// ForceMulti converts MediaOption to MultiMediaOption.
// It can produce unexpected RPC errors. Use carefully.
func ForceMulti(opt MediaOption) MultiMediaOption {
	return multiMediaWrapper{MediaOption: opt}
}
