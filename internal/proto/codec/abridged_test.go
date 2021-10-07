package codec

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/nnqq/td/bin"

	"github.com/stretchr/testify/require"
)

func TestAbridged(t *testing.T) {
	bigHeader := func(l int) (packet []byte) {
		// header + 3 bytes of LE
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], uint32(l>>2))

		packet = append([]byte{127}, buf[0:3]...)
		return
	}

	tests := []struct {
		payloadName string
		testData    func() (payload string, packet []byte)
	}{
		{"Small-4b", func() (payload string, packet []byte) {
			payload = "abcd"
			packet = append([]byte{byte(len(payload) >> 2)}, payload...)
			return
		}},

		{"Medium-124b", func() (payload string, packet []byte) {
			payload = strings.Repeat("a", 124)
			packet = append([]byte{byte(len(payload) >> 2)}, payload...)
			return
		}},

		{"Big-1kb", func() (payload string, packet []byte) {
			payload = strings.Repeat("a", 1024)
			packet = bigHeader(len(payload))
			require.Equal(t, []byte{127, 0, 1, 0}, packet)
			packet = append(packet, payload...)
			return
		}},
	}

	for _, test := range tests {
		payload, packet := test.testData()
		t.Run(test.payloadName, func(t *testing.T) {
			t.Run("Write", func(t *testing.T) {
				b := bytes.NewBuffer(nil)
				err := writeAbridged(b, &bin.Buffer{Buf: []byte(payload)})
				require.NoError(t, err)

				require.Equal(t, packet, b.Bytes())
			})

			t.Run("Read", func(t *testing.T) {
				b := &bin.Buffer{}
				err := readAbridged(bytes.NewReader(packet), b)
				require.NoError(t, err)

				require.Equal(t, payload, string(b.Raw()))
			})
		})
	}
}
