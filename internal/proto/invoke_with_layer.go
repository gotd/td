package proto

import (
	"fmt"

	"github.com/gotd/td/bin"
)

// InvokeWithLayer is invokeWithLayer#da9b0d0d function.
//
// invokeWithLayer#da9b0d0d {X:Type} layer:int query:!X = X;
//
// https://core.telegram.org/method/invokeWithLayer
type InvokeWithLayer struct {
	Layer int
	Query bin.Object
}

// InvokeWithLayerID is TL type id of invokeWithLayer#da9b0d0d.
const InvokeWithLayerID = 0xda9b0d0d

// Encode implements bin.Encoder.
func (i InvokeWithLayer) Encode(b *bin.Buffer) error {
	b.PutID(InvokeWithLayerID)
	b.PutInt(i.Layer)
	return i.Query.Encode(b)
}

// Decode implements bin.Decoder.
func (i InvokeWithLayer) Decode(b *bin.Buffer) (err error) {
	if err := b.ConsumeID(InvokeWithLayerID); err != nil {
		return fmt.Errorf("unable to decode invokeWithLayer#da9b0d0d: %w", err)
	}
	i.Layer, err = b.Int()
	if err != nil {
		return fmt.Errorf("unable to decode invokeWithLayer#da9b0d0d: %w", err)
	}
	if err := i.Query.Decode(b); err != nil {
		return fmt.Errorf("unable to decode invokeWithLayer#da9b0d0d: %w", err)
	}
	return nil
}
