// Package clock abstracts time source.
package clock

import "time"

// Clock is current time source.
type Clock interface {
	Now() time.Time
}
