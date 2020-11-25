package proto

import (
	"testing"

	"github.com/ernado/td/bin"

	"github.com/stretchr/testify/require"
)

func TestEncryptedMessage_Encode(t *testing.T) {
	d := EncryptedMessage{
		EncryptedData: []byte{1, 2, 0x1, 0xff, 0xee},
		MsgKey:        bin.Int128{1, 5, 0, 9},
		AuthKeyID:     101561413,
	}
	b := new(bin.Buffer)
	if err := d.Encode(b); err != nil {
		t.Fatal(err)
	}
	decoded := EncryptedMessage{}
	if err := decoded.Decode(b); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, d, decoded)
	require.Zero(t, b.Len(), "buffer should be consumed")
}
