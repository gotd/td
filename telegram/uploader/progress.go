package uploader

import "context"

// ProgressState represents upload state change.
type ProgressState struct {
	// ID of upload.
	ID int64
	// Name of uploading file.
	Name string
	// Part is an ID of uploaded part.
	Part int
	// PartSize is a size of uploaded part.
	PartSize int
	// Uploaded is a total sum of uploaded bytes.
	Uploaded int64
	// Total is a total size of uploading file.
	// May be equal to -1, in case when Upload created without size (stream upload).
	Total int64
}

// Progress is interface of upload process tracker.
type Progress interface {
	Chunk(ctx context.Context, state ProgressState) error
}
