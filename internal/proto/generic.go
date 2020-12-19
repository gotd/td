package proto

import "github.com/gotd/td/bin"

// TType represents any generic T in TL Schema.
type TType interface {
	bin.Encoder
	bin.Decoder
}
