package bin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPool_GetSize(t *testing.T) {
	sizes := []int{0, 1024}

	for _, size := range sizes {
		a := require.New(t)
		p := NewPool(size)

		b := p.Get()
		a.Empty(b.Buf)
		p.Put(b)

		b = p.GetSize(1024)
		a.Len(b.Buf, 1024)
		p.Put(b)
	}
}
