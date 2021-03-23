// Package file contains file service implementation for tgtest server.
package file

import (
	"github.com/gotd/td/telegram/internal/tgtest"
	"github.com/gotd/td/tg"
)

// Service is a Telegram file service.
type Service struct {
	storage Storage
	// Size of part to use in tg.FileHash.
	hashPartSize int
	// Size of range to return in upload.getFileHashes.
	hashRangeSize int
}

// NewService creates new file Service.
func NewService(storage Storage) *Service {
	return &Service{
		storage: storage,
		// Telegram usually uses this values.
		hashPartSize:  131072,
		hashRangeSize: 10,
	}
}

// WitHashPartSize sets size of part to use in tg.FileHash.
func (m *Service) WitHashPartSize(hashPartSize int) *Service {
	m.hashPartSize = hashPartSize
	return m
}

// WitHashPartSize sets size of range to return in upload.getFileHashes.
func (m *Service) WitHashRangeSize(hashRangeSize int) *Service {
	m.hashRangeSize = hashRangeSize
	return m
}

// OnMessage implements tgtest.Handler.
func (m *Service) OnMessage(server *tgtest.Server, req *tgtest.Request) error {
	id, err := req.Buf.PeekID()
	if err != nil {
		return err
	}

	switch id {
	case tg.UploadGetFileRequestTypeID:
		fileReq := tg.UploadGetFileRequest{}
		if err := fileReq.Decode(req.Buf); err != nil {
			return err
		}

		r, err := m.UploadGetFile(req.RequestCtx, &fileReq)
		if err != nil {
			return err
		}

		return server.SendResult(req, r)
	case tg.UploadGetFileHashesRequestTypeID:
		fileReq := tg.UploadGetFileHashesRequest{}
		if err := fileReq.Decode(req.Buf); err != nil {
			return err
		}

		r, err := m.UploadGetFileHashes(req.RequestCtx, &fileReq)
		if err != nil {
			return err
		}

		return server.SendResult(req, &tg.FileHashVector{Elems: r})
	case tg.UploadSaveFilePartRequestTypeID:
		fileReq := tg.UploadSaveFilePartRequest{}
		if err := fileReq.Decode(req.Buf); err != nil {
			return err
		}

		r, err := m.UploadSaveFilePart(req.RequestCtx, &fileReq)
		if err != nil {
			return err
		}

		return server.SendBool(req, r)
	case tg.UploadSaveBigFilePartRequestTypeID:
		fileReq := tg.UploadSaveBigFilePartRequest{}
		if err := fileReq.Decode(req.Buf); err != nil {
			return err
		}

		r, err := m.UploadSaveBigFilePart(req.RequestCtx, &fileReq)
		if err != nil {
			return err
		}

		return server.SendBool(req, r)
	}
	return nil
}

// Register registers service handlers.
func (m *Service) Register(dispatcher *tgtest.Dispatcher) {
	dispatcher.HandleFunc(tg.UploadGetFileRequestTypeID, m.OnMessage)
	dispatcher.HandleFunc(tg.UploadGetFileHashesRequestTypeID, m.OnMessage)
	dispatcher.HandleFunc(tg.UploadSaveFilePartRequestTypeID, m.OnMessage)
	dispatcher.HandleFunc(tg.UploadSaveBigFilePartRequestTypeID, m.OnMessage)
}
