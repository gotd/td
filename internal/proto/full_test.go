package proto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
)

func testData() (packet, payload []byte) {
	return []byte{
		15, 0, 0, 0, // length
		1, 0, 0, 0, // seqNo
		97, 98, 99, // payload
		78, 214, 109, 148, // crc
	}, []byte("abc")
}

func Test_writeFull(t *testing.T) {
	packet, payload := testData()
	b := bytes.NewBuffer(nil)

	buf := &bin.Buffer{Buf: payload}
	err := writeFull(b, 1, buf)
	require.NoError(t, err)

	require.Equal(t, packet, b.Bytes())
}

func Test_readFull(t *testing.T) {
	packet, payload := testData()

	t.Run("ok", func(t *testing.T) {
		b := &bin.Buffer{}
		err := readFull(bytes.NewBuffer(packet), 1, b)
		require.NoError(t, err)

		require.Equal(t, payload, b.Buf)
	})
}
