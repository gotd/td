package file

import (
	"crypto/rand"
	"sync"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// document is a stored file published as tg.Document.
type document struct {
	name       string
	accessHash int64
	reference  []byte
}

// documents stores published documents by ID.
type documents struct {
	mux  sync.Mutex
	docs map[int64]*document
}

func newReference() []byte {
	ref := make([]byte, 16)
	_, _ = rand.Read(ref)
	return ref
}

// Document publishes stored file (e.g. uploaded file, named
// "<user id>_<file id>") as document with given ID, so it can be downloaded
// via tg.InputDocumentFileLocation with its file reference.
func (m *Service) Document(fileName string, id int64) (*tg.Document, error) {
	f, err := m.storage.Open(fileName)
	if err != nil {
		return nil, errors.Wrap(err, "open")
	}
	size := f.Size()
	_ = f.Close()

	m.docs.mux.Lock()
	defer m.docs.mux.Unlock()
	if m.docs.docs == nil {
		m.docs.docs = map[int64]*document{}
	}
	d, ok := m.docs.docs[id]
	if !ok {
		d = &document{
			name:       fileName,
			accessHash: id ^ 0x5bd1e995,
			reference:  newReference(),
		}
		m.docs.docs[id] = d
	}

	return &tg.Document{
		ID:            id,
		AccessHash:    d.accessHash,
		FileReference: append([]byte(nil), d.reference...),
		Size:          int64(size),
		MimeType:      "application/octet-stream",
		DCID:          2,
	}, nil
}

// ExpireFileReference makes current file reference of document expire:
// requests with it fail with FILE_REFERENCE_EXPIRED until the document is
// fetched again via Document.
func (m *Service) ExpireFileReference(id int64) {
	m.docs.mux.Lock()
	defer m.docs.mux.Unlock()
	if d, ok := m.docs.docs[id]; ok {
		d.reference = newReference()
	}
}

func (m *Service) openDocument(loc *tg.InputDocumentFileLocation) (File, error) {
	m.docs.mux.Lock()
	d, ok := m.docs.docs[loc.ID]
	m.docs.mux.Unlock()
	if !ok || d.accessHash != loc.AccessHash {
		return nil, tgerr.New(400, tg.ErrFileIDInvalid)
	}
	m.docs.mux.Lock()
	valid := string(d.reference) == string(loc.FileReference)
	m.docs.mux.Unlock()
	if !valid {
		return nil, tgerr.New(400, tg.ErrFileReferenceExpired)
	}

	f, err := m.storage.Open(d.name)
	if err != nil {
		return nil, tgerr.New(400, tg.ErrFileIDInvalid)
	}
	return f, nil
}
