package mtproto

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/proto"
)

func gzip(b *bin.Buffer) (*bin.Buffer, error) {
	var content proto.GZIP
	if err := content.Decode(b); err != nil {
		return nil, errors.Wrap(err, "decode")
	}
	return &bin.Buffer{Buf: content.Data}, nil
}

func (c *Conn) handleGZIP(msgID int64, b *bin.Buffer) error {
	content, err := gzip(b)
	if err != nil {
		return errors.Wrap(err, "unzip")
	}
	return c.handleMessage(msgID, content)
}
