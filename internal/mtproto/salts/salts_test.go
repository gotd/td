package salts

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/internal/mt"
)

var testData = []mt.FutureSalt{
	{
		ValidSince: 10,
		ValidUntil: 25,
		Salt:       1,
	},
	{
		ValidSince: 20,
		ValidUntil: 35,
		Salt:       2,
	},
	{
		ValidSince: 30,
		ValidUntil: 45,
		Salt:       3,
	},
}

func TestSalts(t *testing.T) {
	a := require.New(t)
	salts := &Salts{}

	salts.Store(testData[:2])
	a.Len(salts.salts, 2)

	salt, ok := salts.Get(time.Unix(11, 0))
	a.Equal(int64(1), salt)
	a.True(ok)

	_, ok = salts.Get(time.Unix(36, 0))
	a.False(ok)

	salts.Store(testData[:2])
	a.Len(salts.salts, 2)

	salts.Store(testData[:3])
	a.Len(salts.salts, 3)

	salt, ok = salts.Get(time.Unix(26, 0))
	a.Equal(int64(2), salt)
	a.True(ok)

	salt, ok = salts.Get(time.Unix(36, 0))
	a.Equal(int64(3), salt)
	a.True(ok)

	salts.Reset()
	_, ok = salts.Get(time.Unix(36, 0))
	a.False(ok)
}

func BenchmarkSalts_Get(b *testing.B) {
	salts := &Salts{}
	salts.Store(testData)
	t := time.Unix(11, 0)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		salts.Get(t)
	}
}

func BenchmarkSalts_Store(b *testing.B) {
	salts := &Salts{
		salts: make([]mt.FutureSalt, 0, len(testData)),
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		salts.Store(testData)
	}
}
