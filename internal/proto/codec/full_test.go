package codec

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
)

func fullTestData() (packet, payload []byte) {
	return []byte{
		15, 0, 0, 0, // length
		1, 0, 0, 0, // seqNo
		97, 98, 99, // payload
		78, 214, 109, 148, // crc
	}, []byte("abc")
}

func TestFull(t *testing.T) {
	packet, payload := fullTestData()
	t.Run("write", func(t *testing.T) {
		b := bytes.NewBuffer(nil)

		buf := &bin.Buffer{Buf: payload}
		err := writeFull(b, 1, buf)
		require.NoError(t, err)

		require.Equal(t, packet, b.Bytes())
	})

	t.Run("read", func(t *testing.T) {
		b := &bin.Buffer{}
		err := readFull(bytes.NewBuffer(packet), 1, b)
		require.NoError(t, err)

		require.Equal(t, payload, b.Buf)
	})
}
