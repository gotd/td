package dcs

import "github.com/gotd/td/tg"

// List is a list of Telegram DC addresses and domains.
type List struct {
	Options []tg.DCOption
	Domains map[int]string
	Test    bool
}

// Zero returns true if this List is zero value.
func (d List) Zero() bool {
	return d.Options == nil && d.Domains == nil && !d.Test
}
