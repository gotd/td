package tgtest

import (
	"fmt"

	"github.com/nnqq/td/bin"
)

// genericVector is a simple helper to encode a vector of TL objects.
type genericVector struct {
	Elems []bin.Encoder
}

// Encode implements bin.Encoder.
func (vec *genericVector) Encode(b *bin.Buffer) error {
	b.PutVectorHeader(len(vec.Elems))
	for idx, v := range vec.Elems {
		if v == nil {
			return fmt.Errorf("unable to encode Vector<%T>: field Elems element with index %d is nil", v, idx)
		}
		if err := v.Encode(b); err != nil {
			return fmt.Errorf("unable to encode Vector<%T>: field Elems element with index %d: %w", v, idx, err)
		}
	}
	return nil
}
