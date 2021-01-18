package faketls

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_writeClientHello(t *testing.T) {
	var faketlsStartBytes = [...]byte{
		0x16,
		0x03,
		0x01,
		0x02,
		0x00,
		0x01,
		0x00,
		0x01,
		0xfc,
		0x03,
		0x03,
	}

	a := require.New(t)
	b := bytes.NewBuffer(nil)
	_, err := writeClientHello(b, [32]byte{}, make([]byte, 16))
	a.NoError(err)
	a.Equal(faketlsStartBytes[:], b.Bytes()[:len(faketlsStartBytes)])
}
