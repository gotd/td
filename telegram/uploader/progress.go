package uploader

import "context"

// ProgressState represents upload state change.
type ProgressState struct {
	ID       int64
	Name     string
	Part     int
	PartSize int
	Uploaded int
	Total    int
}

// Progress is interface of upload process tracker.
type Progress interface {
	Chunk(ctx context.Context, state ProgressState) error
}
