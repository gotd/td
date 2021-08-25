package tdesktop

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/multierr"
	"golang.org/x/xerrors"
)

type tdesktopFile struct {
	io.Reader
	version uint32
}

func open(tdata, fileName string) (tdesktopFile, error) {
	suffixes := []string{"0", "1", "s"}

	tryRead := func(p string) (_ tdesktopFile, rErr error) {
		f, err := os.Open(p)
		if err != nil {
			return tdesktopFile{}, xerrors.Errorf("open: %w", err)
		}
		defer multierr.AppendInvoke(&rErr, multierr.Close(f))

		return fromFile(f)
	}

	for _, suffix := range suffixes {
		p := filepath.Join(tdata, fileName+suffix)
		if _, err := os.Stat(p); err != nil {
			if os.IsNotExist(err) || os.IsPermission(err) {
				continue
			}
			return tdesktopFile{}, xerrors.Errorf("stat: %w", err)
		}

		f, err := tryRead(p)
		if err != nil {
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
	if magic != [4]byte{'T', 'D', 'F', '$'} {
		return tdesktopFile{}, &WrongMagicError{
			Magic: magic,
		}
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return tdesktopFile{}, xerrors.Errorf("read data: %w", err)
	}
	hash := data[len(data)-16:]
	data = data[:len(data)-16]

	h := md5.New()
	_, _ = h.Write(data)
	var packedLength [4]byte
	binary.LittleEndian.PutUint32(packedLength[:], uint32(len(data)))
	_, _ = h.Write(packedLength[:])
	_, _ = h.Write(version[:])
	_, _ = h.Write(magic[:])

	if !bytes.Equal(h.Sum(nil), hash) {
		return tdesktopFile{}, xerrors.New("hash mismatch")
	}

	v := binary.LittleEndian.Uint32(version[:])
	return tdesktopFile{
		Reader:  bytes.NewReader(data),
		version: v,
	}, nil
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
