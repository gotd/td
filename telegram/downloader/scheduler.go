package downloader

import (
	"context"
	"sync"

	"github.com/gotd/td/tgerr"
)

// SchedulerOptions configures Scheduler.
type SchedulerOptions struct {
	// Min is the minimum window (in-flight part requests). Defaults to 1.
	Min int
	// Start is the initial window. Defaults to 4.
	Start int
	// Max is the maximum window. Defaults to 16.
	Max int
}

func (o *SchedulerOptions) setDefaults() {
	if o.Min <= 0 {
		o.Min = 1
	}
	if o.Max <= 0 {
		o.Max = 16
	}
	if o.Max < o.Min {
		o.Max = o.Min
	}
	if o.Start <= 0 {
		o.Start = 4
	}
	if o.Start < o.Min {
		o.Start = o.Min
	}
	if o.Start > o.Max {
		o.Start = o.Max
	}
}

// SchedulerSnapshot is a snapshot of Scheduler state, e.g. for metrics.
type SchedulerSnapshot struct {
	// Window is the current limit of in-flight part requests.
	Window int
	// InFlight is the number of in-flight part requests.
	InFlight int
	// Waiting is the number of workers waiting for a slot.
	Waiting int
	// Decreases is the total number of window decreases (flood waits and
	// timeouts).
	Decreases int64
}

// Scheduler is an adaptive limit of in-flight part requests shared by
// downloads, like the per-DC download window of Telegram Desktop.
//
// The window grows by one after each window of successful requests and is
// halved on flood wait or timeout, at most once per window of requests in
// flight: concurrent requests failing together halve it once.
//
// Flood waits and timeouts handled inside the CDN download path (after
// upload.fileCdnRedirect) are retried there and are not seen by Scheduler,
// so it limits CDN requests but does not adapt to CDN throttling. Use one Scheduler per DC (or per
// connection pool), and pass it to downloads via Downloader.WithScheduler or
// Builder.WithScheduler; Builder.WithThreads still limits workers of one
// download, so set it high enough to fill the window.
type Scheduler struct {
	mux       sync.Mutex
	opts      SchedulerOptions
	window    int
	inFlight  int
	waiting   int
	successes int
	decreases int64
	// epoch is incremented on each window decrease: a request acquired
	// before the last decrease does not decrease the window again, so a
	// burst of failures of concurrent requests halves it once.
	epoch uint64
	// wake is closed and replaced when a slot may become available.
	wake chan struct{}
}

// NewScheduler creates new Scheduler.
func NewScheduler(opts SchedulerOptions) *Scheduler {
	opts.setDefaults()
	return &Scheduler{
		opts:   opts,
		window: opts.Start,
		wake:   make(chan struct{}),
	}
}

// Snapshot returns current state.
func (s *Scheduler) Snapshot() SchedulerSnapshot {
	s.mux.Lock()
	defer s.mux.Unlock()
	return SchedulerSnapshot{
		Window:    s.window,
		InFlight:  s.inFlight,
		Waiting:   s.waiting,
		Decreases: s.decreases,
	}
}

func (s *Scheduler) notify() {
	close(s.wake)
	s.wake = make(chan struct{})
}

// acquire takes a slot, returning the epoch to pass to release.
func (s *Scheduler) acquire(ctx context.Context) (uint64, error) {
	s.mux.Lock()
	for s.inFlight >= s.window {
		wake := s.wake
		s.waiting++
		s.mux.Unlock()
		select {
		case <-wake:
		case <-ctx.Done():
			s.mux.Lock()
			s.waiting--
			s.mux.Unlock()
			return 0, ctx.Err()
		}
		s.mux.Lock()
		s.waiting--
	}
	s.inFlight++
	epoch := s.epoch
	s.mux.Unlock()
	return epoch, nil
}

// release releases slot taken by acquire, adapting window by request result.
func (s *Scheduler) release(ctx context.Context, epoch uint64, err error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.inFlight--

	switch {
	case err == nil:
		s.successes++
		if s.successes >= s.window && s.window < s.opts.Max {
			s.window++
			s.successes = 0
		}
	case isSchedulerBackoff(ctx, err):
		s.successes = 0
		if epoch != s.epoch {
			// Window was already decreased after this request was sent.
			break
		}
		s.epoch++
		s.decreases++
		s.window /= 2
		if s.window < s.opts.Min {
			s.window = s.opts.Min
		}
	}
	s.notify()
}

func isSchedulerBackoff(ctx context.Context, err error) bool {
	if _, ok := tgerr.AsFloodWait(err); ok {
		return true
	}
	return isRetryableTimeout(ctx, err)
}

// WithScheduler sets Scheduler shared by downloads of this Downloader.
func (d *Downloader) WithScheduler(s *Scheduler) *Downloader {
	d.scheduler = s
	return d
}

// WithScheduler sets Scheduler for this download, overriding the one set via
// Downloader.WithScheduler.
func (b *Builder) WithScheduler(s *Scheduler) *Builder {
	b.scheduler = s
	return b
}
