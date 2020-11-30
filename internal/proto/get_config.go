package proto

import "github.com/gotd/td/bin"

type GetConfig struct{}

func (g GetConfig) Encode(b *bin.Buffer) error {
	b.PutID(0xc4f9186b)
	return nil
}
