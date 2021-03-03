package message

import (
	"context"

	"github.com/gotd/td/tg"
)

type multiMediaBuilder struct {
	sender *Sender
	peer   tg.InputPeerClass
	// Attached media.
	media []tg.InputSingleMedia
}

// MediaOption is a option for sending media attachments.
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

// MultiMediaOption is a option for sending albums.
type MultiMediaOption interface {
	MediaOption
	applyMulti(ctx context.Context, b *multiMediaBuilder) error
}

var _ MultiMediaOption = multiMediaOptionFunc(nil)

// multiMediaOptionFunc is a function adapter for MediaOption.
type multiMediaOptionFunc func(ctx context.Context, b *multiMediaBuilder) error

// apply implements MediaOption.
func (m multiMediaOptionFunc) apply(ctx context.Context, b *multiMediaBuilder) error {
	return m(ctx, b)
}

// applyMulti implements MultiMediaOption.
func (m multiMediaOptionFunc) applyMulti(ctx context.Context, b *multiMediaBuilder) error {
	return m(ctx, b)
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
