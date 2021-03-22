package file

import (
	"context"
	"fmt"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/crypto"
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

func (m *Service) getPart(loc tg.InputFileLocationClass, offset, limit int) ([]byte, error) {
	f, err := m.openLocation(loc)
	if err != nil {
		return nil, err
	}

	r := make([]byte, limit)
	n, err := f.ReadAt(r, int64(offset))
	if err != nil {
		return nil, xerrors.Errorf("read from storage: %w", err)
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

func countHashes(data []byte, partSize int) []tg.FileHash {
	actions := data
	batchSize := partSize
	batches := make([][]byte, 0, (len(actions)+batchSize-1)/batchSize)

	for batchSize < len(actions) {
		actions, batches = actions[batchSize:], append(batches, actions[0:batchSize:batchSize])
	}
	batches = append(batches, actions)

	currentRange := make([]tg.FileHash, 0, 10)
	offset := 0
	for _, batch := range batches {
		currentRange = append(currentRange, tg.FileHash{
			Offset: offset,
			Limit:  partSize,
			Hash:   crypto.SHA256(batch),
		})
		offset += len(batch)
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
// Illustration:
//
// 	Bytes
// 	[0 1 2 3 4 5 6 7 8 9 10 11 ... 111 122 123 124 125 126 127]
//	Parts of bytes (in partSize)
//  [		0			1			2			3			4 ]
// 	Hash range for multiple parts (in rangeSize)
//	[					1					2				  ]
//
func computeBatch(offset, rangeSize, partSize int) int {
	parts := divAndCeil(offset+1, partSize)
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

	if request.Offset >= f.Size() {
		return nil, nil
	}
	// Telegram usually uses this values.
	partSize := 131072
	rangeSize := 10
	batch := computeBatch(request.Offset, rangeSize, partSize)

	low := (batch - 1) * rangeSize * partSize
	high := batch * rangeSize * partSize

	r := make([]byte, high-low)
	_, err = f.ReadAt(r, int64(low))
	if err != nil {
		return nil, err
	}

	return countHashes(r, partSize), nil
}
