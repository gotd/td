package telegram

import (
	"fmt"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/require"
)

func Test_findDC(t *testing.T) {
	options := []tg.DCOption{
		{ID: 1, Ipv6: false},
		{ID: 1, Ipv6: true},
		{ID: 1, Ipv6: false, Static: true},

		{ID: 2, Ipv6: true, Static: true},
		{ID: 2, Ipv6: true},
		{ID: 2, Ipv6: false},
	}
	for i := range options {
		options[i].IPAddress = fmt.Sprintf("DC: %d, Index: %d", options[i].ID, i)
	}
	cfg := tg.Config{DCOptions: options}

	a := require.New(t)
	_, ok := findDC(cfg, -2, false)
	a.False(ok)
	_, ok = findDC(cfg, -2, true)
	a.False(ok)

	// Prefer IPv6.
	dc, ok := findDC(cfg, 1, true)
	a.True(ok)
	a.True(dc.Ipv6)

	// Prefer static.
	dc, ok = findDC(cfg, 1, false)
	a.True(ok)
	a.True(dc.Static)

	// Prefer static and IPv6.
	dc, ok = findDC(cfg, 2, true)
	a.True(ok)
	a.True(dc.Static)
	a.True(dc.Ipv6)
}
