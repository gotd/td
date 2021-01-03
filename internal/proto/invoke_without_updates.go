package proto

import (
	"fmt"

	"github.com/gotd/td/bin"
)

// InvokeWithoutUpdates is invokeWithoutUpdates#bf9459b7 function.
//
// invokeWithoutUpdates#bf9459b7 {X:Type} query:!X = X;
//
// https://core.telegram.org/method/invokeWithoutUpdates
type InvokeWithoutUpdates struct {
	Query bin.Object
}

// InvokeWithoutUpdatesTypeID is TL type id of invokeWithoutUpdates#bf9459b7.
const InvokeWithoutUpdatesTypeID = 0xbf9459b7

// Encode implements bin.Encoder.
func (i InvokeWithoutUpdates) Encode(b *bin.Buffer) error {
	b.PutID(InvokeWithoutUpdatesTypeID)
	return i.Query.Encode(b)
}

// Decode implements bin.Decoder.
func (i InvokeWithoutUpdates) Decode(b *bin.Buffer) (err error) {
	if err := b.ConsumeID(InvokeWithoutUpdatesTypeID); err != nil {
		return fmt.Errorf("unable to decode invokeWithoutUpdates#bf9459b7: %w", err)
	}
	if err := i.Query.Decode(b); err != nil {
		return fmt.Errorf("unable to decode invokeWithoutUpdates#bf9459b7: %w", err)
	}
	return nil
}
