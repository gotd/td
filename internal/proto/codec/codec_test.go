package codec

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
)

type codec interface {
	Write(w io.Writer, b *bin.Buffer) error
	Read(r io.Reader, b *bin.Buffer) error
}

type codecTest struct {
	name   string
	create func() codec
}

func codecs() []codecTest {
	return []codecTest{
		{"Abridged", func() codec {
			return Abridged{}
		}},
		{"Intermediate", func() codec {
			return Intermediate{}
		}},
		{"PaddedIntermediate", func() codec {
			return PaddedIntermediate{}
		}},
		{"Full", func() codec {
			return &Full{}
		}},
	}
}

type payload struct {
	name     string
	testData string
	mustFail bool
}

func payloads() []payload {
	return []payload{
		{"Empty", "", true},
		{"Small 8b", "abcdabcd", false},
		{"Medium 1kb", strings.Repeat("a", 1024), false},
	}
}

func testGood(c codecTest, p payload) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("One message", func(t *testing.T) {
			a := require.New(t)
			codec := c.create()
			buf := bytes.NewBuffer(nil)
			payload := &bin.Buffer{Buf: []byte(p.testData)}

			// Encode
			a.NoError(codec.Write(buf, payload))
			// Decode
			payload.Reset()
			a.NoError(codec.Read(buf, payload))
			a.Equal(p.testData, string(payload.Buf))
		})

		t.Run("Two messages", func(t *testing.T) {
			a := require.New(t)
			codec := c.create()
			buf := bytes.NewBuffer(nil)
			payload := &bin.Buffer{Buf: []byte(p.testData)}

			// Encode twice
			a.NoError(codec.Write(buf, payload))
			payload.ResetTo([]byte(p.testData))
			a.NoError(codec.Write(buf, payload))
			// Decode twice
			payload.Reset()
			a.NoError(codec.Read(buf, payload))
			a.Equal(p.testData, string(payload.Buf))
			payload.Reset()
			a.NoError(codec.Read(buf, payload))
			a.Equal(p.testData, string(payload.Buf))
		})
	}
}

func testBad(c codecTest, p payload) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("write", func(t *testing.T) {
			a := require.New(t)
			payload := &bin.Buffer{Buf: []byte(p.testData)}
			a.Error(c.create().Write(ioutil.Discard, payload))
		})

		t.Run("read", func(t *testing.T) {
			a := require.New(t)
			reader := bytes.NewBufferString(p.testData)
			a.Error(c.create().Read(reader, &bin.Buffer{}))
		})
	}
}

func TestCodecs(t *testing.T) {
	for _, c := range codecs() {
		t.Run(c.name, func(t *testing.T) {
			for _, p := range payloads() {
				if p.mustFail {
					t.Run(p.name, testBad(c, p))
				} else {
					t.Run(p.name, testGood(c, p))
				}
			}
		})
	}
}
