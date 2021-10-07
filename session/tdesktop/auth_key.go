package tdesktop

import (
	"io"
	"math/bits"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/crypto"
)

type reader struct {
	buf bin.Buffer
}

func (r *reader) readUint64() (uint64, error) {
	u, err := r.buf.Uint64()
	return bits.ReverseBytes64(u), err
}

func (r *reader) readUint32() (uint32, error) {
	u, err := r.buf.Uint32()
	return bits.ReverseBytes32(u), err
}

func (r *reader) consumeN(target []byte, n int) error {
	return r.buf.ConsumeN(target, n)
}

func (r *reader) skip(n int) error {
	if r.buf.Len() < n {
		return io.ErrUnexpectedEOF
	}
	r.buf.Skip(n)
	return nil
}

// See https://github.com/telegramdesktop/tdesktop/blob/v2.9.8/Telegram/SourceFiles/storage/storage_account.cpp#L898.
func readMTPData(tgf tdesktopFile, localKey crypto.Key) (MTPAuthorization, error) {
	encrypted, err := tgf.readArray()
	if err != nil {
		return MTPAuthorization{}, xerrors.Errorf("read encrypted data: %w", err)
	}

	decrypted, err := decryptLocal(encrypted, localKey)
	if err != nil {
		return MTPAuthorization{}, xerrors.Errorf("decrypt data: %w", err)
	}
	// Skip decrypted data length (uint32).
	decrypted = decrypted[4:]
	r := reader{buf: bin.Buffer{Buf: decrypted}}

	// TODO(tdakkota): support other IDs.
	var m MTPAuthorization
	if err := m.deserialize(&r); err != nil {
		return MTPAuthorization{}, xerrors.Errorf("deserialize MTPAuthorization: %w", err)
	}
	return m, err
}
