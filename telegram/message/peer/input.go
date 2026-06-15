package peer

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// InputPeer is a peer input accepted by helpers that need a peer.
//
// It is implemented both by concrete tg.InputPeerClass values and by lazy
// references created with Resolve, allowing peers to be passed by
// username/phone/deeplink and resolved lazily.
type InputPeer interface {
	// Zero reports whether the value has a zero value.
	Zero() bool
	// String implements fmt.Stringer.
	String() string
}

// Resolved is a lazily-resolved peer reference produced by Resolve.
//
// It implements InputPeer, so it can be passed directly to helpers that accept
// a peer. The actual resolution is deferred until the helper runs, using the
// resolver provided by that helper.
type Resolved struct {
	fn func(ctx context.Context, r Resolver) (tg.InputPeerClass, error)
}

// Zero reports whether r is unset.
func (r *Resolved) Zero() bool {
	return r == nil || r.fn == nil
}

// String implements fmt.Stringer.
func (r *Resolved) String() string {
	return "peer.Resolved"
}

// Bind binds the given resolver to the reference, producing a Promise.
func (r *Resolved) Bind(resolver Resolver) Promise {
	return func(ctx context.Context) (tg.InputPeerClass, error) {
		if r.Zero() {
			return nil, errors.New("empty peer reference")
		}
		return r.fn(ctx, resolver)
	}
}

// Resolve creates a lazy peer reference from the given text.
//
// It auto-detects domain, phone and deeplink forms. The reference is resolved
// lazily, using the resolver of the helper it is passed to.
// Input examples:
//
//	@telegram
//	telegram
//	t.me/telegram
//	https://t.me/telegram
//	tg:resolve?domain=telegram
//	tg://resolve?domain=telegram
//	+13115552368
//	+1 (311) 555-0123
//	+1 311 555-6162
//	13115556162
func Resolve(from string) *Resolved {
	return &Resolved{fn: func(ctx context.Context, r Resolver) (tg.InputPeerClass, error) {
		return resolve(r, from)(ctx)
	}}
}

// ResolveInputPeer resolves input to a concrete tg.InputPeerClass.
//
// Concrete tg.InputPeerClass values are returned as-is; lazy references created
// by Resolve are resolved using r.
func ResolveInputPeer(ctx context.Context, r Resolver, input InputPeer) (tg.InputPeerClass, error) {
	switch v := input.(type) {
	case nil:
		return nil, errors.New("nil peer")
	case tg.InputPeerClass:
		return v, nil
	case *Resolved:
		return v.Bind(r)(ctx)
	default:
		return nil, errors.Errorf("unsupported peer input type %T", input)
	}
}
