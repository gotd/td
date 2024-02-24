package testutil

import "io"

// ErrWriter returns an io.Writer that returns 0, err from all Write calls.
func ErrWriter(err error) io.Writer {
	return &errWriter{err: err}
}

type errWriter struct {
	err error
}

func (r *errWriter) Write(p []byte) (int, error) {
	return 0, r.err
}
