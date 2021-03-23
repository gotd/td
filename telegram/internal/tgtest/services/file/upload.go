package file

import (
	"context"
	"fmt"

	"go.uber.org/multierr"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/internal/helpers"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

type upload interface {
	// Random file identifier created by the client
	GetFileID() int64
	// Numerical order of a part
	GetFilePart() int
	// Binary data, contend of a part
	GetBytes() []byte
}

func validatePartSize(got, stored int) *tgerr.Error {
	switch {
	case got == 0:
		return tgerr.New(400, tg.ErrFilePartEmpty)
	case got > helpers.MaximumPartSize:
		return tgerr.New(400, tg.ErrFilePartTooBig)
	}

	if stored == 0 {
		return nil
	}

	switch {
	case got != stored:
		return tgerr.New(400, tg.ErrFilePartSizeChanged)
	case helpers.MaximumPartSize%got != 0,
		got%helpers.PaddingPartSize != 0:
		return tgerr.New(400, tg.ErrFilePartSizeInvalid)
	default:
		return nil
	}
}

func (m *Service) write(ctx context.Context, request upload) (err error) {
	// TODO(tdakkota): Better way to handle user id. For now we haven't auth service to pair
	//  user ID and authkey
	id, ok := ctx.Value("user_id").(int)
	if !ok {
		id = 10
	}

	file, err := m.storage.Open(fmt.Sprintf("%d_%d", id, request.GetFileID()))
	if err != nil {
		return xerrors.Errorf("open file: %w", err)
	}
	defer func() {
		multierr.AppendInto(&err, file.Close())
	}()

	part := request.GetFilePart()
	if part < 0 || part > helpers.PartsLimit {
		return tgerr.New(400, tg.ErrFilePartInvalid)
	}
	data := request.GetBytes()
	partSize := file.PartSize()
	if err := validatePartSize(len(data), partSize); err != nil {
		return err
	}
	if partSize == 0 {
		partSize = len(data)
		file.SetPartSize(partSize)
	}

	offset := int64(partSize * part)
	if _, err := file.WriteAt(data, offset); err != nil {
		return xerrors.Errorf("write at %d-%d", offset, offset+int64(len(data)))
	}

	return nil
}

func (m *Service) UploadSaveFilePart(ctx context.Context, request *tg.UploadSaveFilePartRequest) (bool, error) {
	if err := m.write(ctx, request); err != nil {
		return false, err
	}

	return true, nil
}

func (m *Service) UploadSaveBigFilePart(ctx context.Context, request *tg.UploadSaveBigFilePartRequest) (bool, error) {
	part := request.FileTotalParts
	if part < 0 || part > helpers.PartsLimit {
		return false, tgerr.New(400, tg.ErrFilePartsInvalid)
	}

	if err := m.write(ctx, request); err != nil {
		return false, err
	}

	return true, nil
}
