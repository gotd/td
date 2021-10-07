package crypto

import (
	"testing"

	"github.com/nnqq/td/bin"

	"github.com/stretchr/testify/require"
)

func TestEncryptedMessageData_Encode(t *testing.T) {
	const padding = 16

	paddingRequired := func(l int) int {
		return padding + (padding - (l % padding))
	}

	d := EncryptedMessageData{
		Salt:                   1034,
		SeqNo:                  1,
		MessageID:              3401235566,
		SessionID:              2345512351,
		MessageDataLen:         5,
		MessageDataWithPadding: []byte{1, 2, 3, 100, 112},
	}
	b := new(bin.Buffer)
	if err := d.Encode(b); err != nil {
		t.Fatal(err)
	}
	b.Buf = append(b.Buf, make([]byte, paddingRequired(b.Len()))...)
	decoded := EncryptedMessageData{}
	if err := decoded.Decode(b); err != nil {
		t.Fatal(err)
	}
	decoded.MessageDataWithPadding = decoded.MessageDataWithPadding[:decoded.MessageDataLen]
	require.Equal(t, d, decoded)
}
