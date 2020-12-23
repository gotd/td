package codec

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gotd/td/bin"

	"github.com/stretchr/testify/require"
)

func TestAbridged(t *testing.T) {
	tests := []struct {
		payloadName string
		testData    func() (payload string, packet []byte)
	}{
		{"small", func() (payload string, packet []byte) {
			payload = "abcd"
			packet = append([]byte{byte(len(payload) >> 2)}, payload...)
			return
		}},

		{"big", func() (payload string, packet []byte) {
			payload = strings.Repeat("a", 1024)
			// header + 3 bytes of LE 1024
			packet = append([]byte{127, 0, 1, 0}, payload...)
			return
		}},
	}

	for _, test := range tests {
		payload, packet := test.testData()
		t.Run("write-"+test.payloadName, func(t *testing.T) {
			b := bytes.NewBuffer(nil)
			err := writeAbridged(b, &bin.Buffer{Buf: []byte(payload)})
			require.NoError(t, err)

			require.Equal(t, packet, b.Bytes())
		})

		t.Run("read-"+test.payloadName, func(t *testing.T) {
			b := &bin.Buffer{}
			err := readAbridged(bytes.NewReader(packet), b)
			require.NoError(t, err)

			require.Equal(t, payload, string(b.Raw()))
		})
	}
}
