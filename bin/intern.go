//+build !nointern

package bin

import "github.com/josharian/intern"

func strFromBytes(b []byte) string {
	return intern.Bytes(b)
}
