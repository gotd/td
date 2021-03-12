package rpcmock

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
)

// InvokeRaw implements tg.Invoker.
func (i *Mock) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	h := i.Handler()

	body, err := h(input)
	if err != nil {
		return xerrors.Errorf("mock invoke: %w", err)
	}

	buf := new(bin.Buffer)
	if err := body.Encode(buf); err != nil {
		return xerrors.Errorf("encode: %w", err)
	}
	if err := output.Decode(buf); err != nil {
		return xerrors.Errorf("decode: %w", err)
	}
	return nil
}
