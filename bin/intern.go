//+build !nointern

package bin

import "github.com/josharian/intern"

func strFromBytes(b []byte) string {
	// One byte string does not cause allocation.
	if len(b) == 1 {
		return string(b)
	}
	return intern.Bytes(b)
}
