package codec

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/testutil"
)

type codecTest struct {
	name   string
	align  int
	create func() Codec
}

func codecs() []codecTest {
	return []codecTest{
		{"Abridged", 4, func() Codec {
			return Abridged{}
		}},
		{"Intermediate", 4, func() Codec {
			return Intermediate{}
		}},
		{"PaddedIntermediate", 4, func() Codec {
			return PaddedIntermediate{}
		}},
		{"Full", 0, func() Codec {
			return &Full{}
		}},
		{"NoHeaderIntermediate", 0, func() Codec {
			return NoHeader{Codec: Intermediate{}}
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

			l := payload.Len()
			// Encode
			a.NoError(codec.Write(buf, payload))
			a.Equal(l, payload.Len(), "Codec must not change buffer length")

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

			l := payload.Len()
			// Encode twice
			a.NoError(codec.Write(buf, payload))
			payload.ResetTo([]byte(p.testData))
			a.NoError(codec.Write(buf, payload))
			a.Equal(l, payload.Len(), "Codec must not change buffer length")

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
				payload := &bin.Buffer{Buf: []byte(p.testData)}
				require.Error(t, c.create().Write(io.Discard, payload))
			})
		}

		t.Run("Read", func(t *testing.T) {
			reader := bytes.NewBufferString(p.testData)
			require.Error(t, c.create().Read(reader, &bin.Buffer{}))
		})
	}
}

func testHeaderTag(c codecTest) func(t *testing.T) {
	e := io.ErrClosedPipe
	return func(t *testing.T) {
		t.Run("GoodTag", func(t *testing.T) {
			a := require.New(t)
			buf := bytes.NewBuffer(nil)
			a.NoError(c.create().WriteHeader(buf))
			a.NoError(c.create().ReadHeader(buf))
		})

		if tagged, ok := c.create().(TaggedCodec); ok {
			t.Run("ReadError", func(t *testing.T) {
				r := iotest.ErrReader(e)
				require.ErrorIs(t, c.create().ReadHeader(r), e)
			})
			t.Run("WriteError", func(t *testing.T) {
				w := testutil.ErrWriter(e)
				require.ErrorIs(t, c.create().WriteHeader(w), e)
			})
			t.Run("BadTag", func(t *testing.T) {
				tag := tagged.ObfuscatedTag()
				buf := bytes.NewBuffer(tag)
				tag[0] = 0
				require.ErrorIs(t, c.create().ReadHeader(buf), ErrProtocolHeaderMismatch)
			})
		}
	}
}

func testCodec(c codecTest) func(t *testing.T) {
	e := io.ErrClosedPipe

	return func(t *testing.T) {
		t.Run("ReadError", func(t *testing.T) {
			r := iotest.ErrReader(e)
			require.ErrorIs(t, c.create().Read(r, &bin.Buffer{}), e)
		})

		t.Run("WriteError", func(t *testing.T) {
			w := testutil.ErrWriter(e)
			require.ErrorIs(t,
				c.create().Write(
					w,
					&bin.Buffer{Buf: make([]byte, 16)},
				),
				e,
			)
		})

		if c.align != 0 {
			t.Run("AlignError", func(t *testing.T) {
				require.Error(t,
					c.create().Write(
						io.Discard,
						&bin.Buffer{Buf: make([]byte, c.align-1)},
					),
				)
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
			testCodec(c)(t)
		})
	}
}
