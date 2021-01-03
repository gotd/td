package mtproto

import (
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/proto"
)

func (c *Conn) gzip(b *bin.Buffer) (*bin.Buffer, error) {
	return gzip(b)
}

func gzip(b *bin.Buffer) (*bin.Buffer, error) {
	var content proto.GZIP
	if err := content.Decode(b); err != nil {
		return nil, xerrors.Errorf("decode: %w", err)
	}
	return &bin.Buffer{Buf: content.Data}, nil
}

func (c *Conn) handleGZIP(b *bin.Buffer) error {
	content, err := c.gzip(b)
	if err != nil {
		return xerrors.Errorf("unzip: %w", err)
	}
	return c.handleMessage(content)
}
