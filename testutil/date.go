package testutil

import (
	"time"
)

// Date return date for testing.
func Date() time.Time {
	// Using VK birthday as test date.
	return time.Date(2006, 10, 10,
		13, 42, 15,
		34123,
		time.UTC,
	)
}
