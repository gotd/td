package proto

import (
	"io"
	"testing"

	"github.com/gotd/td/bin"

	"github.com/stretchr/testify/require"
)

func TestUnencryptedMessage_Encode(t *testing.T) {
	d := UnencryptedMessage{
		MessageID:   3401235567,
		MessageData: []byte{1, 2, 3, 100, 112},
	}
	b := new(bin.Buffer)
	if err := d.Encode(b); err != nil {
		t.Fatal(err)
	}
	decoded := UnencryptedMessage{}
	if err := decoded.Decode(b); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, d, decoded)
	require.Zero(t, b.Len(), "buffer should be consumed")
}

func TestUnencryptedMessage_DecodeInvalidDataLength(t *testing.T) {
	t.Run("Negative", func(t *testing.T) {
		b := new(bin.Buffer)
		b.PutLong(0)
		b.PutLong(1)
		b.PutInt32(-1)

		var decoded UnencryptedMessage
		err := decoded.Decode(b)
		require.Error(t, err)

		var lengthErr *bin.InvalidLengthError
		require.ErrorAs(t, err, &lengthErr)
		require.Zero(t, decoded.MessageData)
	})

	t.Run("ExceedsRemainingBuffer", func(t *testing.T) {
		b := new(bin.Buffer)
		b.PutLong(0)
		b.PutLong(1)
		b.PutInt32(1024)

		var decoded UnencryptedMessage
		err := decoded.Decode(b)
		require.ErrorIs(t, err, io.ErrUnexpectedEOF)
		require.Zero(t, decoded.MessageData)
	})
}
