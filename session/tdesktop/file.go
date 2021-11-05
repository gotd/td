package tdesktop

import (
	"bytes"
	"crypto/md5" // #nosec G501
	"encoding/binary"
	"io"
	"io/fs"
	"math"

	"github.com/ogen-go/errors"
	"go.uber.org/multierr"
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
			return tdesktopFile{}, errors.Wrap(err, "open")
		}
		defer multierr.AppendInvoke(&rErr, multierr.Close(f))

		return fromFile(f)
	}

	for _, suffix := range suffixes {
		p := fileName + suffix
		if _, err := fs.Stat(filesystem, p); err != nil {
			if errors.Is(err, fs.ErrNotExist) ||
				errors.Is(err, fs.ErrPermission) {
				continue
			}
			return tdesktopFile{}, errors.Wrap(err, "stat")
		}

		f, err := tryRead(p)
		if err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				continue
			}

			var magicErr *WrongMagicError
			if errors.As(err, &magicErr) {
				continue
			}
			return tdesktopFile{}, errors.Wrap(err, "read tdesktop file")
		}

		return f, nil
	}

	return tdesktopFile{}, errors.Errorf("file %q not found", fileName)
}

var tdesktopFileMagic = [4]byte{'T', 'D', 'F', '$'}

// fromFile creates new Telegram Desktop storage file.
// Based on https://github.com/telegramdesktop/tdesktop/blob/v2.9.8/Telegram/SourceFiles/storage/details/storage_file_utilities.cpp#L473.
func fromFile(r io.Reader) (tdesktopFile, error) {
	buf := make([]byte, 16)
	if _, err := io.ReadFull(r, buf[:8]); err != nil {
		return tdesktopFile{}, errors.Wrap(err, "read magic and version")
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
		return tdesktopFile{}, errors.Wrap(err, "read data")
	}
	if l := len(data); l < 16 {
		return tdesktopFile{}, errors.Errorf("invalid data length %d", l)
	}
	hash := data[len(data)-16:]
	data = data[:len(data)-16]

	computedHash := telegramFileHash(data, version)
	if !bytes.Equal(computedHash[:], hash) {
		return tdesktopFile{}, errors.New("hash mismatch")
	}

	v := binary.LittleEndian.Uint32(version[:])
	return tdesktopFile{
		Reader:  bytes.NewBuffer(data),
		version: v,
	}, nil
}

func writeFile(w io.Writer, data []byte, version [4]byte) error {
	if _, err := w.Write(tdesktopFileMagic[:]); err != nil {
		return errors.Wrap(err, "write magic")
	}
	if _, err := w.Write(version[:]); err != nil {
		return errors.Wrap(err, "write version")
	}
	if _, err := w.Write(data); err != nil {
		return errors.Wrap(err, "write data")
	}
	hash := telegramFileHash(data, version)
	if _, err := w.Write(hash[:]); err != nil {
		return errors.Wrap(err, "write hash")
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
		return nil, errors.Wrap(err, "read length")
	}

	// See https://github.com/qt/qtbase/blob/5.15.2/src/corelib/text/qbytearray.cpp#L3314.
	length := order.Uint32(r)
	if length == 0xffffffff {
		return nil, nil
	}

	r = append(r[:0], make([]byte, length)...)
	if _, err := io.ReadFull(reader, r); err != nil {
		return nil, errors.Wrap(err, "read")
	}
	return r, nil
}

func writeArray(writer io.Writer, data []byte, order binary.ByteOrder) error {
	length := len(data)
	if uint64(length) > uint64(math.MaxUint32) {
		return errors.Errorf("data length too big (%d)", length)
	}

	r := make([]byte, 4)
	order.PutUint32(r, uint32(length))
	if _, err := writer.Write(r); err != nil {
		return errors.Wrap(err, "write length")
	}

	if _, err := writer.Write(data); err != nil {
		return errors.Wrap(err, "write data")
	}

	return nil
}
