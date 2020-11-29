// +build !fuzz

package telegram

type Zero struct{}

func (Zero) Read(p []byte) (n int, err error) { return len(p), nil }
