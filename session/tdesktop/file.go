package tdesktop

import (
	"bytes"
	"crypto/md5" // #nosec G501
	"encoding/binary"
	"io"
	"io/fs"
	"math"

	"go.uber.org/multierr"
	"golang.org/x/xerrors"
)

type tdesktopFile struct {
	io.Reader
	version uint32
}

func open(filesystem fs.FS, fileName string) (tdesktopFile, error) {
	suffixes := []string{"0", "1", "s"}

	tryRead := func(p string) (_ tdesktopFile, rErr error) {
		f, err := filesystem.Open(p)
		if err != nil {
			return tdesktopFile{}, xerrors.Errorf("open: %w", err)
		}
		defer multierr.AppendInvoke(&rErr, multierr.Close(f))

		return fromFile(f)
	}

	for _, suffix := range suffixes {
		p := fileName + suffix
		if _, err := fs.Stat(filesystem, p); err != nil {
			if xerrors.Is(err, fs.ErrNotExist) ||
				xerrors.Is(err, fs.ErrPermission) {
				continue
			}
			return tdesktopFile{}, xerrors.Errorf("stat: %w", err)
		}

		f, err := tryRead(p)
		if err != nil {
			if xerrors.Is(err, io.ErrUnexpectedEOF) {
				continue
			}

			var magicErr *WrongMagicError
			if xerrors.As(err, &magicErr) {
				continue
			}
			return tdesktopFile{}, xerrors.Errorf("read tdesktop file: %w", err)
		}

		return f, nil
	}

	return tdesktopFile{}, xerrors.Errorf("file %q not found", fileName)
}

var tdesktopFileMagic = [4]byte{'T', 'D', 'F', '$'}

// fromFile creates new Telegram Desktop storage file.
// Based on https://github.com/telegramdesktop/tdesktop/blob/v2.9.8/Telegram/SourceFiles/storage/details/storage_file_utilities.cpp#L473.
func fromFile(r io.Reader) (tdesktopFile, error) {
	buf := make([]byte, 16)
	if _, err := io.ReadFull(r, buf[:8]); err != nil {
		return tdesktopFile{}, xerrors.Errorf("read magic and version: %w", err)
	}

	var magic, version [4]byte
	copy(magic[:], buf[:4])
	// TODO(tdakkota): check version
	copy(version[:], buf[4:8])
	if magic != tdesktopFileMagic {
		return tdesktopFile{}, &WrongMagicError{
			Magic: magic,
		}
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return tdesktopFile{}, xerrors.Errorf("read data: %w", err)
	}
	if l := len(data); l < 16 {
		return tdesktopFile{}, xerrors.Errorf("invalid data length %d", l)
	}
	hash := data[len(data)-16:]
	data = data[:len(data)-16]

	computedHash := telegramFileHash(data, version)
	if !bytes.Equal(computedHash[:], hash) {
		return tdesktopFile{}, xerrors.New("hash mismatch")
	}

	v := binary.LittleEndian.Uint32(version[:])
	return tdesktopFile{
		Reader:  bytes.NewBuffer(data),
		version: v,
	}, nil
}

func writeFile(w io.Writer, data []byte, version [4]byte) error {
	if _, err := w.Write(tdesktopFileMagic[:]); err != nil {
		return xerrors.Errorf("write magic: %w", err)
	}
	if _, err := w.Write(version[:]); err != nil {
		return xerrors.Errorf("write version: %w", err)
	}
	if _, err := w.Write(data); err != nil {
		return xerrors.Errorf("write data: %w", err)
	}
	hash := telegramFileHash(data, version)
	if _, err := w.Write(hash[:]); err != nil {
		return xerrors.Errorf("write hash: %w", err)
	}
	return nil
}

func telegramFileHash(data []byte, version [4]byte) (r [md5.Size]byte) {
	h := md5.New() // #nosec G401
	_, _ = h.Write(data)
	var packedLength [4]byte
	binary.LittleEndian.PutUint32(packedLength[:], uint32(len(data)))
	_, _ = h.Write(packedLength[:])
	_, _ = h.Write(version[:])
	_, _ = h.Write(tdesktopFileMagic[:])
	h.Sum(r[:0])
	return r
}

func (f tdesktopFile) readArray() ([]byte, error) {
	return readArray(f.Reader, binary.BigEndian)
}

func readArray(reader io.Reader, order binary.ByteOrder) ([]byte, error) {
	r := make([]byte, 32)
	if _, err := io.ReadFull(reader, r[:4]); err != nil {
		return nil, xerrors.Errorf("read length: %w", err)
	}

	// See https://github.com/qt/qtbase/blob/5.15.2/src/corelib/text/qbytearray.cpp#L3314.
	length := order.Uint32(r)
	if length == 0xffffffff {
		return nil, nil
	}

	r = append(r[:0], make([]byte, length)...)
	if _, err := io.ReadFull(reader, r); err != nil {
		return nil, xerrors.Errorf("read: %w", err)
	}
	return r, nil
}

func writeArray(writer io.Writer, data []byte, order binary.ByteOrder) error {
	length := len(data)
	if length > math.MaxUint32 {
		return xerrors.Errorf("data length too big (%d)", length)
	}

	r := make([]byte, 4)
	order.PutUint32(r, uint32(length))
	if _, err := writer.Write(r); err != nil {
		return xerrors.Errorf("write length: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		return xerrors.Errorf("write data: %w", err)
	}

	return nil
}
