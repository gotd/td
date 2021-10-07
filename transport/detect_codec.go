package transport

import (
	"bytes"
	"io"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/proto/codec"
)

func detectCodec(c io.Reader) (Codec, io.Reader, error) {
	var buf [4]byte
	if _, err := io.ReadFull(c, buf[:1]); err != nil {
		return nil, nil, xerrors.Errorf("read first byte: %w", err)
	}

	if buf[0] == codec.AbridgedClientStart[0] {
		return Abridged.Codec(), c, nil
	}

	if _, err := io.ReadFull(c, buf[1:4]); err != nil {
		return nil, nil, xerrors.Errorf("read header: %w", err)
	}
	switch buf {
	case codec.IntermediateClientStart:
		return Intermediate.Codec(), c, nil
	case codec.PaddedIntermediateClientStart:
		return PaddedIntermediate.Codec(), c, nil
	default:
		buffered := bytes.NewReader(buf[:])
		r := io.MultiReader(buffered, c)
		return Full.Codec(), r, nil
	}
}
