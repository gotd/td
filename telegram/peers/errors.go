package peers

import (
	"fmt"

	"github.com/gotd/td/tg"
)

// PhoneNotFoundError is returned when Manager unable to find contact with given phone.
type PhoneNotFoundError struct {
	Phone string
}

// Error implements error.
func (c *PhoneNotFoundError) Error() string {
	return "contact not found"
}

// PeerNotFoundError is returned when Manager unable to find Peer with given tg.PeerClass.
type PeerNotFoundError struct {
	Peer tg.PeerClass
}

// Error implements error.
func (p *PeerNotFoundError) Error() string {
	return fmt.Sprintf("can't resolve %v", p.Peer)
}
