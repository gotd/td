package telegram

import (
	"golang.org/x/xerrors"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/internal/proto"
)

func (c *Client) gzip(b *bin.Buffer) (*bin.Buffer, error) {
	var content proto.GZIP
	if err := content.Decode(b); err != nil {
		return nil, xerrors.Errorf("failed to decode: %w", err)
	}
	return &bin.Buffer{Buf: content.Data}, nil
}

func (c *Client) handleGZIP(b *bin.Buffer) error {
	content, err := c.gzip(b)
	if err != nil {
		return xerrors.Errorf("failed to unzip: %w", err)
	}
	return c.handleMessage(content)
}
