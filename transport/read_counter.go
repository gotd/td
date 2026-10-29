package transport

// ReadCounter is implemented by connections that can report the total number
// of bytes read from the underlying network connection.
type ReadCounter interface {
	// BytesRead returns total bytes read from the underlying connection.
	BytesRead() uint64
}
