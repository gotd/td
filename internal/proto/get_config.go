package proto

import "github.com/gotd/td/bin"

// GetConfig is help.getConfig#c4f9186b function.
type GetConfig struct{}

// Encode implements bin.Encoder.
func (GetConfig) Encode(b *bin.Buffer) error {
	b.PutID(0xc4f9186b)
	return nil
}

// Decode implements bin.Decoder.
func (GetConfig) Decode(b *bin.Buffer) error {
	return b.ConsumeID(0xc4f9186b)
}
