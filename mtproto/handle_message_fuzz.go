// +build fuzz

package mtproto

import (
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/internal/rpc"
)

type Zero struct{}

func (Zero) Read(p []byte) (n int, err error) { return len(p), nil }

func FuzzHandleMessage(data []byte) int {
	c := &Client{
		rand: Zero{},
		rpc:  rpc.New(rpc.NopSend, rpc.Config{}),
		log:  zap.NewNop(),

		sessionCreated: createCondOnce(),
	}
	if err := c.handleMessage(&bin.Buffer{Buf: data}); err != nil {
		return 0
	}
	return 1
}
