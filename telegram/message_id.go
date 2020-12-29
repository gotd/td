package telegram

import "github.com/gotd/td/internal/proto"

func (c *Client) newMessageID() int64 {
	return c.messageID.New(proto.MessageFromClient)
}
