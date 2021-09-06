//go:build !js
// +build !js

package crypto

import (
	"crypto/rand"
	"io"
)

// DefaultRand returns default entropy source.
func DefaultRand() io.Reader {
	return rand.Reader
}
