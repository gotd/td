package file

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func getLocation(loc tg.InputFileLocationClass) (string, error) {
	v, ok := loc.(interface {
		GetLocalID() int
		GetVolumeID() int64
	})
	if !ok {
		return "", tgerr.New(400, tg.ErrFileIDInvalid)
	}

	return fmt.Sprintf("%d_%d", v.GetLocalID(), v.GetVolumeID()), nil
}

func (m *Service) openLocation(loc tg.InputFileLocationClass) (File, error) {
	name, err := getLocation(loc)
	if err != nil {
		return nil, err
	}

	f, err := m.storage.Open(name)
	if err != nil {
		return nil, tgerr.New(400, tg.ErrFileIDInvalid)
	}

	return f, nil
}

func (m *Service) getPart(loc tg.InputFileLocationClass, offset int64, limit int) ([]byte, error) {
	f, err := m.openLocation(loc)
	if err != nil {
		return nil, err
	}

	r := make([]byte, limit)
	n, err := f.ReadAt(r, offset)
	if err != nil {
		return nil, errors.Wrap(err, "read from storage")
	}

	return r[:n], nil
}

func (m *Service) UploadGetFile(ctx context.Context, request *tg.UploadGetFileRequest) (tg.UploadFileClass, error) {
	data, err := m.getPart(request.Location, request.Offset, request.Limit)
	if err != nil {
		return nil, err
	}

	return &tg.UploadFile{
		Type:  &tg.StorageFilePartial{},
		Mtime: 0,
		Bytes: data,
	}, nil
}

func countHashes(data []byte, offset int64, partSize int) []tg.FileHash {
	actions := data
	batchSize := partSize
	batches := make([][]byte, 0, (len(actions)+batchSize-1)/batchSize)

	for batchSize < len(actions) {
		actions, batches = actions[batchSize:], append(batches, actions[0:batchSize:batchSize])
	}
	batches = append(batches, actions)

	currentRange := make([]tg.FileHash, 0, 10)
	for _, batch := range batches {
		currentRange = append(currentRange, tg.FileHash{
			Offset: offset,
			Limit:  partSize,
			Hash:   crypto.SHA256(batch),
		})
		offset += int64(len(batch))
	}
	return currentRange
}

func divAndCeil(a, b int) int {
	r := a / b
	if a%b != 0 {
		r++
	}

	return r
}

// computeBatch computes hash range number for given offset.
func computeBatch(offset int64, rangeSize, partSize int) int {
	// Compute number of parts in partSize from offset.
	parts := divAndCeil(int(offset+1), partSize)
	// Compute number of hash ranges in rangeSize.
	batches := divAndCeil(parts, rangeSize)

	return batches
}

func (m *Service) UploadGetFileHashes(
	ctx context.Context,
	request *tg.UploadGetFileHashesRequest,
) ([]tg.FileHash, error) {
	f, err := m.openLocation(request.Location)
	if err != nil {
		return nil, err
	}

	if request.Offset >= int64(f.Size()) {
		return nil, nil
	}
	partSize := m.hashPartSize
	rangeSize := m.hashRangeSize
	batch := computeBatch(request.Offset, rangeSize, partSize)

	low := (batch - 1) * rangeSize * partSize
	high := batch * rangeSize * partSize

	r := make([]byte, high-low)
	n, err := f.ReadAt(r, int64(low))
	if err != nil {
		return nil, err
	}
	r = r[:n]

	return countHashes(r, int64(low), partSize), nil
}
