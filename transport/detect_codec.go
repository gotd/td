package transport

import (
	"bytes"
	"io"

	"github.com/go-faster/errors"

	"github.com/gotd/td/proto/codec"
)

func detectCodec(c io.Reader) (Codec, io.Reader, error) {
	var buf [4]byte
	if _, err := io.ReadFull(c, buf[:1]); err != nil {
		return nil, nil, errors.Wrap(err, "read first byte")
	}

	if buf[0] == codec.AbridgedClientStart[0] {
		return Abridged.Codec(), c, nil
	}

	if _, err := io.ReadFull(c, buf[1:4]); err != nil {
		return nil, nil, errors.Wrap(err, "read header")
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
