package crypto

import (
	"io"

	"github.com/gotd/td/bin"
)

// RandInt64n returns random int64 from randSource in [0; n).
func RandInt64n(randSource io.Reader, n int64) (int64, error) {
	v, err := RandInt64(randSource)
	if err != nil {
		return 0, err
	}
	if v < 0 {
		v *= -1
	}
	return v % n, nil
}

// RandInt64 returns random int64 from randSource.
func RandInt64(randSource io.Reader) (int64, error) {
	var buf [bin.Word * 2]byte
	if _, err := io.ReadFull(randSource, buf[:]); err != nil {
		return 0, err
	}
	b := &bin.Buffer{Buf: buf[:]}
	return b.Long()
}

// RandInt128 generates and returns new random 128-bit integer.
//
// Use crypto/rand.Reader as randSource in production.
func RandInt128(randSource io.Reader) (bin.Int128, error) {
	var buf [bin.Word * 4]byte
	if _, err := io.ReadFull(randSource, buf[:]); err != nil {
		return bin.Int128{}, err
	}
	b := &bin.Buffer{Buf: buf[:]}
	return b.Int128()
}

// RandInt256 generates and returns new random 256-bit integer.
//
// Use crypto/rand.Reader as randSource in production.
func RandInt256(randSource io.Reader) (bin.Int256, error) {
	var buf [bin.Word * 8]byte
	if _, err := io.ReadFull(randSource, buf[:]); err != nil {
		return bin.Int256{}, err
	}
	b := &bin.Buffer{Buf: buf[:]}
	return b.Int256()
}
