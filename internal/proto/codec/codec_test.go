package codec

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
)

type codecTest struct {
	name   string
	create func() Codec
}

func codecs() []codecTest {
	return []codecTest{
		{"Abridged", func() Codec {
			return Abridged{}
		}},
		{"Intermediate", func() Codec {
			return Intermediate{}
		}},
		{"PaddedIntermediate", func() Codec {
			return PaddedIntermediate{}
		}},
		{"Full", func() Codec {
			return &Full{}
		}},
	}
}

type payload struct {
	name         string
	testData     string
	mustFail     bool
	readTestOnly bool
}

func payloads() []payload {
	var code [4]byte
	binary.LittleEndian.PutUint32(code[:], CodeTransportFlood)
	return []payload{
		{"Empty", "", true, false},
		{"Protocol error", string(code[:]), true, true},
		{"Small 8b", "abcdabcd", false, false},
		{"Medium 1kb", strings.Repeat("a", 1024), false, false},
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
		if !p.readTestOnly {
			t.Run("Write", func(t *testing.T) {
				a := require.New(t)
				payload := &bin.Buffer{Buf: []byte(p.testData)}
				a.Error(c.create().Write(io.Discard, payload))
			})
		}

		t.Run("Read", func(t *testing.T) {
			a := require.New(t)
			reader := bytes.NewBufferString(p.testData)
			a.Error(c.create().Read(reader, &bin.Buffer{}))
		})
	}
}

func testHeaderTag(c codecTest) func(t *testing.T) {
	return func(t *testing.T) {
		skipBad := false
		t.Run("Good tag", func(t *testing.T) {
			a := require.New(t)
			buf := bytes.NewBuffer(nil)
			a.NoError(c.create().WriteHeader(buf))
			if buf.Len() == 0 {
				skipBad = true
			}
			a.NoError(c.create().ReadHeader(buf))
		})

		if !skipBad {
			t.Run("Bad tag", func(t *testing.T) {
				a := require.New(t)
				buf := bytes.NewBuffer(nil)
				a.Error(c.create().ReadHeader(buf))
			})
		}
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

			t.Run("Header", testHeaderTag(c))
		})
	}
}
