package manager

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestFindDC(t *testing.T) {
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
	_, ok := FindDC(cfg, -2, false)
	a.False(ok)
	_, ok = FindDC(cfg, -2, true)
	a.False(ok)

	// Prefer IPv6.
	dc, ok := FindDC(cfg, 1, true)
	a.True(ok)
	a.True(dc.Ipv6)

	// Prefer static.
	dc, ok = FindDC(cfg, 1, false)
	a.True(ok)
	a.True(dc.Static)

	// Prefer static and IPv6.
	dc, ok = FindDC(cfg, 2, true)
	a.True(ok)
	a.True(dc.Static)
	a.True(dc.Ipv6)
}

func TestFindPrimaryDC(t *testing.T) {
	options := []tg.DCOption{
		{ID: 1, Ipv6: false},
		{ID: 1, Ipv6: true},
		{ID: 1, Ipv6: false, Static: true},

		{ID: 2, Ipv6: true, Static: true, MediaOnly: true},
		{ID: 2, Ipv6: true, CDN: true},
		{ID: 2, Ipv6: false, TCPObfuscatedOnly: true},
	}
	for i := range options {
		options[i].IPAddress = fmt.Sprintf("DC: %d, Index: %d", options[i].ID, i)
	}
	cfg := tg.Config{DCOptions: options}
	a := require.New(t)
	_, err := FindPrimaryDC(cfg, -2, false)
	a.Error(err)
	_, err = FindPrimaryDC(cfg, -2, true)
	a.Error(err)

	// Prefer IPv6.
	dc, err := FindPrimaryDC(cfg, 1, true)
	a.NoError(err)
	a.True(dc.Ipv6)

	// Prefer static.
	dc, err = FindPrimaryDC(cfg, 1, false)
	a.NoError(err)
	a.True(dc.Static)

	// Filter CDN/MediaOnly/TCPo.
	dc, err = FindPrimaryDC(cfg, 2, false)
	a.Error(err)
}
