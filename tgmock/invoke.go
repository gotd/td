package tgmock

import (
	"context"
	"crypto/rand"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/crypto"
)

// Invoke implements tg.Invoker.
func (i *Mock) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	h := i.Handler()

	id, err := crypto.RandInt64(rand.Reader)
	if err != nil {
		return xerrors.Errorf("generate id: %w", err)
	}

	body, err := h(id, input)
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
