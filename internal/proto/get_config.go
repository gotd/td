package proto

import "github.com/gotd/td/bin"

// GetConfig is help.getConfig#c4f9186b function.
type GetConfig struct{}

// Encode implements bin.Encoder.
func (GetConfig) Encode(b *bin.Buffer) error {
	b.PutID(0xc4f9186b)
	return nil
}
