package proto

import "github.com/gotd/td/bin"

type InvokeWithLayer struct {
	Layer int
	Query bin.Encoder
}

func (i InvokeWithLayer) Encode(b *bin.Buffer) error {
	b.PutID(0xda9b0d0d)
	b.PutInt(i.Layer)
	return i.Query.Encode(b)
}
