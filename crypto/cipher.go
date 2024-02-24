package crypto

import "io"

// Cipher is message encryption utility struct.
type Cipher struct {
	rand        io.Reader
	encryptSide Side
}

// Rand returns random generator.
func (c Cipher) Rand() io.Reader {
	return c.rand
}

// NewClientCipher creates new client-side Cipher.
func NewClientCipher(rand io.Reader) Cipher {
	return Cipher{rand: rand, encryptSide: Client}
}

// NewServerCipher creates new server-side Cipher.
func NewServerCipher(rand io.Reader) Cipher {
	return Cipher{rand: rand, encryptSide: Server}
}
