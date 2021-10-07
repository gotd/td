package mtproto

import (
	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/proto"
)

func gzip(b *bin.Buffer) (*bin.Buffer, error) {
	var content proto.GZIP
	if err := content.Decode(b); err != nil {
		return nil, xerrors.Errorf("decode: %w", err)
	}
	return &bin.Buffer{Buf: content.Data}, nil
}

func (c *Conn) handleGZIP(msgID int64, b *bin.Buffer) error {
	content, err := gzip(b)
	if err != nil {
		return xerrors.Errorf("unzip: %w", err)
	}
	return c.handleMessage(msgID, content)
}
