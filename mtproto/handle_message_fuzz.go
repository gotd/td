// +build fuzz

package mtproto

import (
	"time"

	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/internal/rpc"
)

type Zero struct{}

func (Zero) Read(p []byte) (n int, err error) { return len(p), nil }

func FuzzHandleMessage(data []byte) int {
	c := &Conn{
		rand:      Zero{},
		rpc:       rpc.New(rpc.NopSend, rpc.Config{}),
		log:       zap.NewNop(),
		messageID: proto.NewMessageIDGen(time.Now, 1),
	}
	if err := c.handleMessage(&bin.Buffer{Buf: data}); err != nil {
		return 0
	}
	return 1
}
