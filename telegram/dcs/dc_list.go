package dcs

import "github.com/gotd/td/tg"

// DCList is a list of Telegram DC addresses and domains.
type DCList struct {
	Options []tg.DCOption
	Domains map[int]string
}

// Zero returns true if this DCList is zero value.
func (d DCList) Zero() bool {
	return d.Options == nil && d.Domains == nil
}
