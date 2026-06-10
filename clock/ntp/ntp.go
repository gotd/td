// Package ntp implements a clock.Clock backed by an NTP-calibrated time source.
//
// It periodically (or on demand) queries an NTP server to estimate the offset
// between the local system clock and the server's clock, and reports a
// corrected time via Now. Timer and Ticker behave the same as clock.System.
//
// See https://core.telegram.org/mtproto#time-synchronization.
package ntp

import (
	"sync"
	"time"

	"github.com/beevik/ntp"
	"github.com/go-faster/errors"

	"github.com/gotd/td/clock"
)

// DefaultServer is the NTP server used when Options.Server is empty.
const DefaultServer = "pool.ntp.org"

// Options configures the network Clock.
type Options struct {
	// Server is the NTP server address to query.
	//
	// Defaults to DefaultServer.
	Server string
	// Now is the local time source the offset is applied to.
	//
	// Defaults to time.Now.
	Now func() time.Time

	// query is the NTP query function, overridable in tests.
	query func(server string) (time.Duration, error)
}

func (o *Options) setDefaults() {
	if o.Server == "" {
		o.Server = DefaultServer
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.query == nil {
		o.query = queryNTP
	}
}

// queryNTP queries the NTP server and returns the estimated clock offset.
func queryNTP(server string) (time.Duration, error) {
	resp, err := ntp.Query(server)
	if err != nil {
		return 0, errors.Wrap(err, "query")
	}
	if err := resp.Validate(); err != nil {
		return 0, errors.Wrap(err, "validate")
	}
	return resp.ClockOffset, nil
}

// Clock is a clock.Clock that corrects the local time using an offset estimated
// from an NTP server.
//
// It is safe for concurrent use.
type Clock struct {
	server string
	now    func() time.Time
	query  func(server string) (time.Duration, error)

	mux    sync.RWMutex
	offset time.Duration
}

var _ clock.Clock = (*Clock)(nil)

// New creates a network Clock and performs the initial synchronization.
func New(opts Options) (*Clock, error) {
	opts.setDefaults()
	c := &Clock{
		server: opts.Server,
		now:    opts.Now,
		query:  opts.query,
	}
	if err := c.Sync(); err != nil {
		return nil, err
	}
	return c, nil
}

// Sync queries the NTP server and updates the recorded clock offset.
func (c *Clock) Sync() error {
	offset, err := c.query(c.server)
	if err != nil {
		return errors.Wrap(err, "sync")
	}
	c.mux.Lock()
	c.offset = offset
	c.mux.Unlock()
	return nil
}

// Offset returns the last recorded offset between the NTP server clock and the
// local clock.
func (c *Clock) Offset() time.Duration {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.offset
}

// Now returns the corrected current time.
func (c *Clock) Now() time.Time {
	c.mux.RLock()
	offset := c.offset
	c.mux.RUnlock()
	return c.now().Add(offset)
}

// Timer returns a timer, behaving the same as clock.System.
func (c *Clock) Timer(d time.Duration) clock.Timer {
	return clock.System.Timer(d)
}

// Ticker returns a ticker, behaving the same as clock.System.
func (c *Clock) Ticker(d time.Duration) clock.Ticker {
	return clock.System.Ticker(d)
}
