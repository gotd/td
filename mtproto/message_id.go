package mtproto

import (
	"github.com/gotd/td/internal/proto"
)

func (c *Client) newMessageID() int64 {
	return int64(proto.NewMessageID(c.clock(), proto.MessageFromClient))
}
