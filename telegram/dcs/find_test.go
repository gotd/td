package dcs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func TestFindDCs(t *testing.T) {
	dcOptions := []tg.DCOption{
		{ID: 1, Ipv6: false},
		{ID: 1, Ipv6: true},
		{ID: 1, Ipv6: false, Static: true},

		{ID: 2, Ipv6: true, Static: true},
		{ID: 2, Ipv6: true},
		{ID: 2, Ipv6: false},
	}
	for i := range dcOptions {
		dcOptions[i].IPAddress = fmt.Sprintf("DC: %d, Index: %d", dcOptions[i].ID, i)
	}

	a := require.New(t)
	dc := FindDCs(dcOptions, -2, false)
	a.Empty(dc)
	dc = FindDCs(dcOptions, -2, true)
	a.Empty(dc)

	// Prefer IPv6.
	dc = FindDCs(dcOptions, 1, true)
	a.True(dc[0].Ipv6)

	// Prefer static.
	dc = FindDCs(dcOptions, 1, false)
	a.True(dc[0].Static)

	// Prefer static and IPv6.
	dc = FindDCs(dcOptions, 2, true)
	a.True(dc[0].Static)
	a.True(dc[0].Ipv6)
}

func TestFindPrimaryDCs(t *testing.T) {
	dcOptions := []tg.DCOption{
		{ID: 1, Ipv6: false},
		{ID: 1, Ipv6: true},
		{ID: 1, Ipv6: false, Static: true},

		{ID: 2, Ipv6: true, Static: true, MediaOnly: true},
		{ID: 2, Ipv6: true, CDN: true},
		{ID: 2, Ipv6: false, TCPObfuscatedOnly: true},
	}
	for i := range dcOptions {
		dcOptions[i].IPAddress = fmt.Sprintf("DC: %d, Index: %d", dcOptions[i].ID, i)
	}
	a := require.New(t)
	dc := FindPrimaryDCs(dcOptions, -2, false)
	a.Empty(dc)
	dc = FindPrimaryDCs(dcOptions, -2, true)
	a.Empty(dc)

	// Prefer IPv6.
	dc = FindPrimaryDCs(dcOptions, 1, true)
	a.True(dc[0].Ipv6)

	// Prefer static.
	dc = FindPrimaryDCs(dcOptions, 1, false)
	a.True(dc[0].Static)

	// Filter CDN/MediaOnly/TCPo.
	dc = FindPrimaryDCs(dcOptions, 2, false)
	a.Empty(dc)
}
