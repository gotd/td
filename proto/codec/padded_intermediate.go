package codec

import (
	"io"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
)

// PaddedIntermediateClientStart is starting bytes sent by client in Padded intermediate mode.
//
// Note that server does not respond with it.
var PaddedIntermediateClientStart = [4]byte{0xdd, 0xdd, 0xdd, 0xdd}

// PaddedIntermediate is intermediate MTProto transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#padded-intermediate
type PaddedIntermediate struct{}

var (
	_ TaggedCodec = PaddedIntermediate{}
)

// WriteHeader sends protocol tag.
func (i PaddedIntermediate) WriteHeader(w io.Writer) error {
	if _, err := w.Write(PaddedIntermediateClientStart[:]); err != nil {
		return errors.Wrap(err, "write padded intermediate header")
	}

	return nil
}

// ReadHeader reads protocol tag.
func (i PaddedIntermediate) ReadHeader(r io.Reader) error {
	var b [4]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return errors.Wrap(err, "read padded intermediate header")
	}

	if b != PaddedIntermediateClientStart {
		return ErrProtocolHeaderMismatch
	}

	return nil
}

// ObfuscatedTag returns protocol tag for obfuscation.
func (i PaddedIntermediate) ObfuscatedTag() [4]byte {
	return PaddedIntermediateClientStart
}

// Write encode to writer message from given buffer.
func (i PaddedIntermediate) Write(w io.Writer, b *bin.Buffer) error {
	if err := checkOutgoingMessage(b); err != nil {
		return err
	}

	if err := checkAlign(b, 4); err != nil {
		return err
	}

	if err := writePaddedIntermediate(crypto.DefaultRand(), w, b); err != nil {
		return errors.Wrap(err, "write padded intermediate")
	}

	return nil
}

// Read fills buffer with received message.
func (i PaddedIntermediate) Read(r io.Reader, b *bin.Buffer) error {
	if err := readPaddedIntermediate(r, b); err != nil {
		return errors.Wrap(err, "read padded intermediate")
	}

	return checkProtocolError(b)
}

func writePaddedIntermediate(randSource io.Reader, w io.Writer, b *bin.Buffer) error {
	length := b.Len()

	b.Expand(4)
	defer func() {
		b.Buf = b.Buf[:length]
	}()

	_, err := io.ReadFull(randSource, b.Buf[length:length+4])
	if err != nil {
		return err
	}
	n := int(b.Buf[length-1]) % 4
	b.Buf = b.Buf[:length+n]

	return writeIntermediate(w, b)
}

func readPaddedIntermediate(r io.Reader, b *bin.Buffer) error {
	if err := readIntermediate(r, b, true); err != nil {
		return err
	}

	padding := b.Len() % 4
	b.Buf = b.Buf[:b.Len()-padding]
	return nil
}
