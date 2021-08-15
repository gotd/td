package codec

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoHeader(t *testing.T) {
	a := require.New(t)
	cdc := NoHeader{
		Codec: Intermediate{},
	}

	buf := bytes.Buffer{}
	a.NoError(cdc.WriteHeader(&buf))
	a.Equal(0, buf.Len())
	a.NoError(cdc.ReadHeader(&buf))
}
