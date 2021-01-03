package mtproto

import "github.com/gotd/td/internal/proto"

func (c *Conn) newMessageID() int64 {
	return c.messageID.New(proto.MessageFromClient)
}
