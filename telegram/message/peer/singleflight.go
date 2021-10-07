package peer

import (
	"context"

	"golang.org/x/sync/singleflight"

	"github.com/nnqq/td/tg"
)

type singleFlight struct {
	next Resolver
	sg   singleflight.Group
}

func (s *singleFlight) ResolveDomain(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	ch := s.sg.DoChan(domain, func() (interface{}, error) {
		return s.next.ResolveDomain(ctx, domain)
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-ch:
		if r.Err != nil {
			return nil, r.Err
		}
		return r.Val.(tg.InputPeerClass), nil
	}
}

func (s *singleFlight) ResolvePhone(ctx context.Context, phone string) (tg.InputPeerClass, error) {
	ch := s.sg.DoChan(phone, func() (interface{}, error) {
		return s.next.ResolvePhone(ctx, phone)
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-ch:
		if r.Err != nil {
			return nil, r.Err
		}
		return r.Val.(tg.InputPeerClass), nil
	}
}

// SingleflightResolver is a simple resolver decorator
// which prevents duplicate resolve calls.
func SingleflightResolver(next Resolver) Resolver {
	return &singleFlight{next: next}
}
