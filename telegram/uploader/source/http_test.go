package source

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHTTPSource(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		a := require.New(t)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		data := bytes.Repeat([]byte{1}, 10)
		var h http.HandlerFunc = func(w http.ResponseWriter, req *http.Request) {
			_, err := w.Write(data)
			a.NoError(err)
		}

		s := httptest.NewServer(h)
		defer s.Close()

		src := new(HTTPSource).WithClient(s.Client())
		f, err := src.Open(ctx, &url.URL{
			Scheme: "http",
			Host:   s.Listener.Addr().String(),
			Path:   "img.jpg",
		})
		a.NoError(err)
		a.Len(data, int(f.Size()))
		a.Equal("img.jpg", f.Name())

		r, err := io.ReadAll(f)
		a.NoError(err)
		a.Equal(data, r)

		a.NoError(f.Close())
	})

	t.Run("NotFound", func(t *testing.T) {
		a := require.New(t)
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		var h http.HandlerFunc = func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}

		s := httptest.NewServer(h)
		defer s.Close()

		src := new(HTTPSource).WithClient(s.Client())
		_, err := src.Open(ctx, &url.URL{
			Scheme: "http",
			Host:   s.Listener.Addr().String(),
			Path:   "img.jpg",
		})
		a.Error(err)
	})
}
