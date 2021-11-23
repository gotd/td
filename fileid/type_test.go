package fileid

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestType_String(t *testing.T) {
	for i := Thumbnail; i <= lastType+1; i++ {
		require.NotEmpty(t, i.String())
	}
}
