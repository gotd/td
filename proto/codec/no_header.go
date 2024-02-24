package codec

import "io"

// NoHeader wraps codec to skip WriteHeader.
type NoHeader struct {
	Codec
}

// WriteHeader implements Codec.
func (NoHeader) WriteHeader(io.Writer) error {
	return nil
}

// ReadHeader implements Codec.
func (NoHeader) ReadHeader(io.Reader) error {
	return nil
}
