package tgtest

import (
	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/tg"
)

// UnpackInvoke is a simple Handler middleware to unpack some Invoke*-like requests.
// Including:
//
//	tg.InvokeWithLayerRequest
// 	tg.InitConnectionRequest
//	tg.InvokeWithoutUpdatesRequest
//
func UnpackInvoke(next Handler) Handler {
	return HandlerFunc(func(srv *Server, req *Request) error {
		id, err := req.Buf.PeekID()
		if err != nil {
			return err
		}

		// TODO(tdakkota): handle more Invoke* requests.
		var (
			obj = peekIDObject{}
			r   bin.Decoder
		)
		for {
			switch id {
			case tg.InvokeWithLayerRequestTypeID:
				r = &tg.InvokeWithLayerRequest{
					Query: &obj,
				}
				// TODO(tdakkota): pass Layer to session.
			case tg.InitConnectionRequestTypeID:
				r = &tg.InitConnectionRequest{
					Query: &obj,
				}
				// TODO(tdakkota): pass DeviceInfo to session.
			case tg.InvokeWithoutUpdatesRequestTypeID:
				r = &tg.InvokeWithoutUpdatesRequest{
					Query: &obj,
				}
				// TODO(tdakkota): pass NoUpdates flag to session.
			default:
				return next.OnMessage(srv, req)
			}

			if err := r.Decode(req.Buf); err != nil {
				return err
			}
			id = obj.TypeID
		}
	})
}

type peekIDObject struct {
	TypeID uint32
}

func (t *peekIDObject) Decode(b *bin.Buffer) error {
	id, err := b.PeekID()
	if err != nil {
		return xerrors.Errorf("peek id: %w", err)
	}
	t.TypeID = id
	return nil
}

func (t *peekIDObject) Encode(*bin.Buffer) error {
	return xerrors.New("peekIDObject must not be encoded")
}
