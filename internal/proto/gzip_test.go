package proto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
)

func TestGZIP_Encode(t *testing.T) {
	data := bytes.Repeat([]byte{1, 2, 3}, 100)
	g := &GZIP{
		Data: data,
	}

	var b bin.Buffer
	require.NoError(t, b.Encode(g))

	var decoded GZIP
	require.NoError(t, b.Decode(&decoded))
	require.Equal(t, data, decoded.Data)
}

func TestGZIP_Decode(t *testing.T) {
	g := &GZIP{
		Data: make([]byte, 1024*1024*15),
	}
	var b bin.Buffer
	require.NoError(t, b.Encode(g))

	var decoded GZIP
	// TODO(ernado): fail explicitly if limit is reached
	require.NoError(t, b.Decode(&decoded))
	require.Less(t, len(decoded.Data), len(g.Data))
}
