//go:build !fuzz
// +build !fuzz

package mtproto

type Zero struct{}

func (Zero) Read(p []byte) (n int, err error) { return len(p), nil }
