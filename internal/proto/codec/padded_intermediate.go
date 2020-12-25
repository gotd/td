package codec

import (
	"crypto/rand"
	"io"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
)

// PaddedIntermediateClientStart is starting bytes sent by client in Padded intermediate mode.
//
// Note that server does not respond with it.
var PaddedIntermediateClientStart = [4]byte{0xdd, 0xdd, 0xdd, 0xdd}

// PaddedIntermediate is intermediate MTProto transport.
//
// See https://core.telegram.org/mtproto/mtproto-transports#padded-intermediate
type PaddedIntermediate struct{}

// WriteHeader sends protocol tag.
func (i PaddedIntermediate) WriteHeader(w io.Writer) error {
	if _, err := w.Write(PaddedIntermediateClientStart[:]); err != nil {
		return xerrors.Errorf("write padded intermediate header: %w", err)
	}

	return nil
}

// ReadHeader reads protocol tag.
func (i PaddedIntermediate) ReadHeader(r io.Reader) error {
	var b [4]byte
	if _, err := r.Read(b[:]); err != nil {
		return xerrors.Errorf("read padded intermediate header: %w", err)
	}

	if b != PaddedIntermediateClientStart {
		return ErrProtocolHeaderMismatch
	}

	return nil
}

// ObfuscatedTag returns protocol tag for obfuscation.
func (i PaddedIntermediate) ObfuscatedTag() (r [4]byte) {
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

	if err := writePaddedIntermediate(rand.Reader, w, b); err != nil {
		return xerrors.Errorf("write padded intermediate: %w", err)
	}

	return nil
}

// Read fills buffer with received message.
func (i PaddedIntermediate) Read(r io.Reader, b *bin.Buffer) error {
	if err := readPaddedIntermediate(r, b); err != nil {
		return xerrors.Errorf("read padded intermediate: %w", err)
	}

	if err := checkProtocolError(b); err != nil {
		return err
	}

	return nil
}

func writePaddedIntermediate(randSource io.Reader, w io.Writer, b *bin.Buffer) error {
	length := b.Len()

	n, err := crypto.RandInt64n(randSource, 4)
	if err != nil {
		return err
	}
	b.Expand(int(n))

	_, err = io.ReadFull(randSource, b.Buf[length:length+int(n)])
	if err != nil {
		return err
	}

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
