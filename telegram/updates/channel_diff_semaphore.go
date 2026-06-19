package updates

import "context"

// chDiffSem bounds how many updates.getChannelDifference requests may be in
// flight at once across all channel workers of a single manager. It is a
// counting semaphore backed by a buffered channel.
//
// A nil chDiffSem imposes no limit, so the manager's default behaviour (every
// channel recovers its gap independently) is preserved with zero overhead.
type chDiffSem chan struct{}

// newChDiffSem returns a semaphore admitting at most n concurrent holders, or
// nil (unlimited) when n is not positive.
func newChDiffSem(n int) chDiffSem {
	if n <= 0 {
		return nil
	}
	return make(chDiffSem, n)
}

// acquire takes a slot, blocking until one is free or ctx is done. It returns
// ctx.Err() if the context is cancelled before a slot is obtained. A nil
// semaphore acquires instantly.
func (s chDiffSem) acquire(ctx context.Context) error {
	if s == nil {
		return nil
	}
	select {
	case s <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// release returns a previously acquired slot. It is a no-op on a nil semaphore
// and must be called exactly once per successful acquire.
func (s chDiffSem) release() {
	if s == nil {
		return
	}
	<-s
}
