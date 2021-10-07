package dcs

import "github.com/nnqq/td/tg"

// Prod returns production DC list.
func Prod() List {
	return List{
		Options: []tg.DCOption{
			{
				ID:        1,
				IPAddress: "149.154.175.59",
				Port:      443,
			},
			{
				Static:    true,
				ID:        1,
				IPAddress: "149.154.175.53",
				Port:      443,
			},
			{
				Ipv6:      true,
				ID:        1,
				IPAddress: "2001:0b28:f23d:f001:0000:0000:0000:000a",
				Port:      443,
			},
			{
				ID:        2,
				IPAddress: "149.154.167.50",
				Port:      443,
			},
			{
				Static:    true,
				ID:        2,
				IPAddress: "149.154.167.51",
				Port:      443,
			},
			{
				MediaOnly: true,
				ID:        2,
				IPAddress: "149.154.167.151",
				Port:      443,
			},
			{
				Ipv6:      true,
				ID:        2,
				IPAddress: "2001:067c:04e8:f002:0000:0000:0000:000a",
				Port:      443,
			},
			{
				Ipv6:      true,
				MediaOnly: true,
				ID:        2,
				IPAddress: "2001:067c:04e8:f002:0000:0000:0000:000b",
				Port:      443,
			},
			{
				Static:    true,
				ID:        3,
				IPAddress: "149.154.175.100",
				Port:      443,
			},
			{
				Ipv6:      true,
				ID:        3,
				IPAddress: "2001:0b28:f23d:f003:0000:0000:0000:000a",
				Port:      443,
			},
			{
				Static:    true,
				ID:        4,
				IPAddress: "149.154.167.91",
				Port:      443,
			},
			{
				Ipv6:      true,
				ID:        4,
				IPAddress: "2001:067c:04e8:f004:0000:0000:0000:000a",
				Port:      443,
			},
			{
				MediaOnly: true,
				ID:        4,
				IPAddress: "149.154.166.120",
				Port:      443,
			},
			{
				Ipv6:      true,
				MediaOnly: true,
				ID:        4,
				IPAddress: "2001:067c:04e8:f004:0000:0000:0000:000b",
				Port:      443,
			},
			{
				Ipv6:      true,
				ID:        5,
				IPAddress: "2001:0b28:f23f:f005:0000:0000:0000:000a",
				Port:      443,
			},
			{
				Static:    true,
				ID:        5,
				IPAddress: "91.108.56.173",
				Port:      443,
			},
		},
		Domains: map[int]string{
			1: "wss://pluto.web.telegram.org/apiws",
			2: "wss://venus.web.telegram.org/apiws",
			3: "wss://aurora.web.telegram.org/apiws",
			4: "wss://vesta.web.telegram.org/apiws",
			5: "wss://flora.web.telegram.org/apiws",
		},
	}
}
