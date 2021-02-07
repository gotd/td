// Package uploader contains uploading files helpers.
package uploader

import (
	"io"

	"go.uber.org/atomic"
)

// NewUpload creates new Upload struct using given
// name and reader.
func NewUpload(name string, from io.Reader, total int64) *Upload {
	return &Upload{
		name:       name,
		totalBytes: total,
		from:       from,
		partSize:   -1,
	}
}

// FromReader creates new Upload struct using
// given io.Reader.
func FromReader(name string, from io.Reader, total int64) *Upload {
	return &Upload{
		name:       name,
		totalBytes: total,
		from:       from,
		partSize:   -1,
	}
}

// Upload represents Telegram file upload.
type Upload struct {
	// Fields which will be set by Uploader.
	// File ID for Telegram.
	id int64
	// Sent parts (in partSize).
	sentParts atomic.Int64
	// Confirmed uploaded parts.
	confirmedParts atomic.Int64
	// Confirmed uploaded bytes.
	confirmedBytes atomic.Int64
	// Total parts.
	totalParts int
	// Part size of uploader.
	partSize int
	// Flag to determine class of size of file.
	big bool

	// Total size (in bytes) of upload.
	totalBytes int64 // immutable
	// Name of file.
	name string // immutable
	// Reader of data.
	from io.Reader // immutable
}

func (u *Upload) confirmSmall(bytes int) ProgressState {
	part := int(u.confirmedParts.Inc())
	return u.confirm(part, bytes)
}

func (u *Upload) confirm(part, bytes int) ProgressState {
	uploaded := int(u.confirmedBytes.Add(int64(bytes)))

	return ProgressState{
		ID:       u.id,
		Name:     u.name,
		Part:     part,
		PartSize: u.partSize,
		Uploaded: uploaded,
		Total:    int(u.totalBytes),
	}
}
