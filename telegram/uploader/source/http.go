package source

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/go-faster/errors"
	"go.uber.org/multierr"
)

// HTTPSource is HTTP source.
type HTTPSource struct {
	client *http.Client
}

// NewHTTPSource creates new HTTPSource.
func NewHTTPSource() *HTTPSource {
	return &HTTPSource{client: http.DefaultClient}
}

// WithClient sets HTTP client to use.
func (s *HTTPSource) WithClient(client *http.Client) *HTTPSource {
	s.client = client
	return s
}

type httpFile struct {
	body io.ReadCloser
	name string
	size int64
}

func (h httpFile) Read(p []byte) (n int, err error) {
	return h.body.Read(p)
}

func (h httpFile) Close() error {
	return h.body.Close()
}

func (h httpFile) Name() string {
	return h.name
}

func (h httpFile) Size() int64 {
	return h.size
}

// Open implements Source.
func (s *HTTPSource) Open(ctx context.Context, u *url.URL) (_ RemoteFile, rerr error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "create request")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}
	defer func() {
		if rerr != nil {
			multierr.AppendInto(&rerr, resp.Body.Close())
		}
	}()
	if resp.StatusCode >= 400 {
		return nil, errors.Errorf("bad code %d", resp.StatusCode)
	}

	lastURL := u
	if resp.Request.URL != nil {
		lastURL = resp.Request.URL
	}

	return httpFile{
		body: resp.Body,
		name: path.Base(lastURL.Path),
		size: resp.ContentLength,
	}, nil
}
