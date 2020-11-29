// +build fuzz

package telegram

import (
	"go.uber.org/zap"

	"github.com/ernado/td/bin"
)

type Zero struct{}

func (Zero) Read(p []byte) (n int, err error) { return len(p), nil }

func FuzzHandleMessage(data []byte) int {
	c := &Client{
		rand: Zero{},
		log:  zap.NewNop(),
	}
	if err := c.handleMessage(&bin.Buffer{Buf: data}); err != nil {
		return 0
	}
	return 1
}
