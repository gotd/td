package transport

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/internal/proto/codec"
)

func Test_detectCodec(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		resultType Codec
		wantErr    bool
	}{
		{
			"Abridged",
			codec.AbridgedClientStart[:],
			codec.Abridged{},
			false,
		},
		{
			"Intermediate",
			codec.IntermediateClientStart[:],
			codec.Intermediate{}, false,
		},
		{
			"PaddedIntermediate",
			codec.PaddedIntermediateClientStart[:],
			codec.PaddedIntermediate{},
			false,
		},
		{
			"Full",
			[]byte{'g', 'o', 't', 'd'},
			&codec.Full{},
			false,
		},
		{
			"EOF-first",
			nil,
			nil,
			true,
		},
		{
			"EOF-second",
			[]byte{'a'},
			nil,
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := require.New(t)

			r, _, err := detectCodec(bytes.NewReader(test.data))
			if test.wantErr {
				a.Nil(r)
				a.Error(err)
			} else {
				a.NotNil(r)
				a.NoError(err)
				a.IsType(test.resultType, r)
			}
		})
	}
}
