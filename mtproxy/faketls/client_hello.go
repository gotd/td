package faketls

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"io"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
)

const clientHelloLength = 517

func createClientHello(b *bin.Buffer, sessionID [32]byte, domain string, key [32]byte) (randomOffset int) {
	S := func(s string) {
		b.Buf = append(b.Buf, s...)
	}
	Z := func(n int) {
		randomOffset = len(b.Buf)
		b.Expand(n)
	}
	G := func(_ int) {
		b.Expand(0)
	}
	R := func() {
		b.Buf = append(b.Buf, sessionID[:]...)
	}
	D := func() {
		b.Buf = append(b.Buf, domain...)
	}
	K := func() {
		b.Buf = append(b.Buf, key[:]...)
	}
	var stack []int
	Open := func() {
		stack = append(stack, b.Len())
		b.Expand(2)
	}
	Close := func() {
		lastIdx := len(stack) - 1
		s := stack[lastIdx]
		stack = stack[:lastIdx]

		length := b.Len() - (s + 2)
		binary.BigEndian.PutUint16(b.Buf[s:], uint16(length))
	}

	S("\x16\x03\x01\x02\x00\x01\x00\x01\xfc\x03\x03")
	Z(32)
	S("\x20")
	R()
	S("\x00\x20")
	G(0)
	S("\x13\x01\x13\x02\x13\x03\xc0\x2b\xc0\x2f\xc0\x2c\xc0\x30\xcc\xa9" +
		"\xcc\xa8\xc0\x13\xc0\x14\x00\x9c\x00\x9d\x00\x2f\x00\x35\x01\x00" +
		"\x01\x93")
	G(2)
	S("\x00\x00\x00\x00")
	Open()
	Open()
	S("\x00")
	Open()
	D()
	Close()
	Close()
	Close()
	S("\x00\x17\x00\x00\xff\x01\x00\x01\x00\x00\x0a\x00\x0a\x00\x08")
	G(4)
	S("\x00\x1d\x00\x17\x00\x18\x00\x0b\x00\x02\x01\x00\x00\x23\x00\x00" +
		"\x00\x10\x00\x0e\x00\x0c\x02\x68\x32\x08\x68\x74\x74\x70\x2f\x31" +
		"\x2e\x31\x00\x05\x00\x05\x01\x00\x00\x00\x00\x00\x0d\x00\x12\x00" +
		"\x10\x04\x03\x08\x04\x04\x01\x05\x03\x08\x05\x05\x01\x08\x06\x06" +
		"\x01\x00\x12\x00\x00\x00\x33\x00\x2b\x00\x29")
	G(4)
	S("\x00\x01\x00\x00\x1d\x00\x20")
	K()
	S("\x00\x2d\x00\x02\x01\x01\x00\x2b\x00\x0b\x0a")
	G(6)
	S("\x03\x04\x03\x03\x03\x02\x03\x01\x00\x1b\x00\x03\x02\x00\x02")
	G(3)
	S("\x00\x01\x00\x00\x15")

	if pad := clientHelloLength - b.Len(); pad > 0 {
		b.Expand(pad)
	}
	return randomOffset
}

// writeClientHello writes faketls ClientHello.
//
// See https://tools.ietf.org/html/rfc5246#section-7.4.1.1.
func writeClientHello(
	w io.Writer,
	now clock.Clock,
	sessionID [32]byte,
	domain string,
	secret []byte,
) (r [32]byte, err error) {
	b := &bin.Buffer{
		Buf: make([]byte, 0, 576),
	}
	randomOffset := createClientHello(b, sessionID, domain, [32]byte{})

	// https://github.com/tdlib/td/blob/27d3fdd09d90f6b77ecbcce50b1e86dc4b3dd366/td/mtproto/TlsInit.cpp#L380-L384
	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write(b.Buf); err != nil {
		return [32]byte{}, errors.Wrap(err, "hmac write")
	}

	s := mac.Sum(nil)
	copy(b.Buf[randomOffset:randomOffset+32], s)
	// Overwrite last 4 bytes using final := original ^ timestamp.
	old := binary.LittleEndian.Uint32(b.Buf[randomOffset+28 : randomOffset+32])
	old ^= uint32(now.Now().Unix())
	binary.LittleEndian.PutUint32(b.Buf[randomOffset+28:randomOffset+32], old)

	// Copy ClientRandom for later use.
	copy(r[:], b.Buf[randomOffset:randomOffset+32])
	_, err = w.Write(b.Buf)
	return r, err
}
