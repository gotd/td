package faketls

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/go-faster/errors"
)

const maxTLSRecordDataLength = 16384 + 24

type record struct {
	Type    RecordType
	Version [2]byte
	Data    []byte
}

func readRecord(r io.Reader) (record, error) {
	rec := record{}

	buf := make([]byte, 5)
	if _, err := io.ReadFull(r, buf); err != nil {
		return record{}, err
	}

	rec.Type = RecordType(buf[0])
	versionRaw := buf[1:3]
	switch {
	case bytes.Equal(versionRaw, Version13Bytes[:]):
		rec.Version = Version13Bytes
	case bytes.Equal(versionRaw, Version12Bytes[:]):
		rec.Version = Version12Bytes
	case bytes.Equal(versionRaw, Version11Bytes[:]):
		rec.Version = Version11Bytes
	case bytes.Equal(versionRaw, Version10Bytes[:]):
		rec.Version = Version10Bytes
	default:
		return record{}, errors.Errorf("unknown protocol version %v", versionRaw)
	}

	length := binary.BigEndian.Uint16(buf[3:])
	if length > maxTLSRecordDataLength {
		return record{}, errors.New("record length is too big")
	}

	rec.Data = make([]byte, length)
	if _, err := io.ReadFull(r, rec.Data); err != nil {
		return record{}, err
	}

	return rec, nil
}

func writeRecord(w io.Writer, r record) (int, error) {
	buf := [...]byte{
		byte(r.Type),
		r.Version[0], r.Version[1],
	}

	if _, err := w.Write(buf[:]); err != nil {
		return 0, errors.Wrap(err, "type and version")
	}

	binary.BigEndian.PutUint16(buf[:2], uint16(len(r.Data)))
	if _, err := w.Write(buf[:2]); err != nil {
		return 0, errors.Wrap(err, "data length")
	}

	n, err := w.Write(r.Data)
	if err != nil {
		return 0, errors.Wrap(err, "data")
	}

	return n, nil
}
