package tgmock

import (
	"context"

	"github.com/nnqq/td/bin"
)

// Invoker implements tg.Invoker as function.
type Invoker func(request bin.Encoder) (bin.Encoder, error)

// Invoke implements tg.Invoker.
func (f Invoker) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	resp, err := f(input)
	if err != nil {
		return err
	}
	b := &bin.Buffer{}
	if err := resp.Encode(b); err != nil {
		return err
	}
	return output.Decode(b)
}
