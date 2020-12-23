package codec

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
)

type codec interface {
	Write(w io.Writer, b *bin.Buffer) error
	Read(r io.Reader, b *bin.Buffer) error
}

func codecs() []struct {
	name   string
	create func() codec
} {
	return []struct {
		name   string
		create func() codec
	}{
		{"abridged", func() codec {
			return Abridged{}
		}},
		{"intermediate", func() codec {
			return Intermediate{}
		}},
		{"full", func() codec {
			return &Full{}
		}},
	}
}

func payloads() []struct {
	payloadName string
	payload     string
} {
	return []struct {
		payloadName string
		payload     string
	}{
		{"small-8b", "abcdabcd"},
		{"medium-1kb", strings.Repeat("a", 1024)},
	}
}

func TestCodecs(t *testing.T) {
	for _, c := range codecs() {
		for _, test := range payloads() {
			t.Run(fmt.Sprintf("%s:%s", c.name, test.payloadName), func(t *testing.T) {
				a := require.New(t)
				codec := c.create()
				buf := bytes.NewBuffer(nil)

				// Encode
				payload := &bin.Buffer{Buf: []byte(test.payload)}
				a.NoError(codec.Write(buf, payload))
				// Decode
				payload.Reset()
				a.NoError(codec.Read(buf, payload))
				a.Equal(test.payload, string(payload.Buf))
			})
		}
	}
}
