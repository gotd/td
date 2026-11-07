package downloader

import (
	"context"
	"io"

	"github.com/go-faster/errors"
)

// Verify checks local file data (e.g. a partially downloaded .part file)
// against file hashes from the server (upload.getFileHashes) and returns the
// length of the longest verified prefix.
//
// The r must contain the first size bytes of the file. A trailing part
// shorter than a hash window can be verified only if it is the end of the
// file, otherwise it is not counted.
//
// Use the result with WithOffset to resume a download safely: if
// verification is enabled, the offset only has to be 4KB-aligned, otherwise
// round it down to the part size.
func (b *Builder) Verify(ctx context.Context, r io.ReaderAt, size int64) (int64, error) {
	if size <= 0 {
		return 0, nil
	}

	v := newVerifier(b.schema, b.hashes...)
	var (
		verified int64
		buf      []byte
	)
	for {
		hash, ok, err := v.next(ctx)
		if err != nil {
			return verified, errors.Wrap(err, "get hashes")
		}
		if !ok || hash.Offset != verified || hash.Offset >= size {
			// No more hashes, unexpected gap, or end of local data.
			return verified, nil
		}

		n := int64(hash.Limit)
		if hash.Offset+n > size {
			n = size - hash.Offset
		}
		if int64(cap(buf)) < n {
			buf = make([]byte, n)
		}
		data := buf[:n]
		if _, err := r.ReadAt(data, hash.Offset); err != nil && !errors.Is(err, io.EOF) {
			return verified, errors.Wrap(err, "read")
		}
		if !v.verify(hash, data) {
			// Mismatch, or partial window that is not the end of the file.
			return verified, nil
		}

		verified = hash.Offset + n
		if n < int64(hash.Limit) {
			// Last (short) window of the file.
			return verified, nil
		}
	}
}
