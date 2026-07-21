package telegram

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestClientForwardsPingOptions pins the telegram.Options -> mtproto.Options
// forwarding of the ping/idle tuning knobs inside NewClient's mtproto.Options{}
// literal. These four fields were silently dropped between the two Options
// structs for the whole life of the package, and nothing in ./telegram/...
// noticed because the assignment site had zero test coverage. Deleting any
// one assignment in client.go must fail exactly the corresponding case below.
func TestClientForwardsPingOptions(t *testing.T) {
	tests := []struct {
		name string
		set  func(*Options, time.Duration)
		get  func(*Client) time.Duration
	}{
		{
			name: "PingInterval",
			set:  func(o *Options, d time.Duration) { o.PingInterval = d },
			get:  func(c *Client) time.Duration { return c.opts.PingInterval },
		},
		{
			name: "PingTimeout",
			set:  func(o *Options, d time.Duration) { o.PingTimeout = d },
			get:  func(c *Client) time.Duration { return c.opts.PingTimeout },
		},
		{
			name: "PingDelayDisconnect",
			set:  func(o *Options, d time.Duration) { o.PingDelayDisconnect = d },
			get:  func(c *Client) time.Duration { return c.opts.PingDelayDisconnect },
		},
		{
			name: "IdleTimeout",
			set:  func(o *Options, d time.Duration) { o.IdleTimeout = d },
			get:  func(c *Client) time.Duration { return c.opts.IdleTimeout },
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Distinctive, mutually-different, non-default value per field so
			// a mixup between two fields (not just a dropped assignment)
			// would also be caught.
			want := time.Duration(100+i) * time.Second

			var opt Options
			tt.set(&opt, want)

			c := NewClient(1, "hash", opt)

			require.Equal(t, want, tt.get(c))
		})
	}
}
