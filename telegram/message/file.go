package message

import "github.com/nnqq/td/tg"

// FileLocation is an abstraction of Telegram file location.
type FileLocation interface {
	GetID() (value int64)
	GetAccessHash() (value int64)
	GetFileReference() (value []byte)
}

func inputDocuments(files ...FileLocation) (r []tg.InputDocumentClass) {
	r = make([]tg.InputDocumentClass, len(files))
	for i := range files {
		v := new(tg.InputDocument)
		v.FillFrom(files[i])
		r[i] = v
	}

	return
}
