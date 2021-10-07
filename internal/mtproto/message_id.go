package mtproto

import "github.com/nnqq/td/internal/proto"

func (c *Conn) newMessageID() int64 {
	return c.messageID.New(proto.MessageFromClient)
}
