package proto

import "github.com/gotd/td/bin"

// InvokeWithLayer is invokeWithLayer#da9b0d0d function.
//
// invokeWithLayer#da9b0d0d {X:Type} layer:int query:!X = X;
//
// https://core.telegram.org/method/invokeWithLayer
type InvokeWithLayer struct {
	Layer int
	Query bin.Encoder
}

// Encode implements bin.Encoder.
func (i InvokeWithLayer) Encode(b *bin.Buffer) error {
	b.PutID(0xda9b0d0d)
	b.PutInt(i.Layer)
	return i.Query.Encode(b)
}
