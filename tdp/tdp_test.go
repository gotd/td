package tdp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormat(t *testing.T) {
	for _, tt := range []struct {
		Input  interface{}
		Output string
	}{
		{Output: "<nil>"},
	} {
		t.Run(tt.Output, func(t *testing.T) {
			require.Equal(t, tt.Output, Format(tt.Input))
		})
	}
}
