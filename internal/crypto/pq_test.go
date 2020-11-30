package crypto

import (
	"math/big"
	"math/rand"
	"testing"
)

func TestDecomposePQ(t *testing.T) {
	for _, tt := range []struct {
		pq, p, q *big.Int
	}{
		// Testing vectors taken from one of SplitPQ implementations.
		// * https://github.com/sdidyk/mtproto/blob/cf2cb57ade6932a7e0854f2e3246492a2028d369/math_test.go
		{big.NewInt(1724114033281923457), big.NewInt(1229739323), big.NewInt(1402015859)},
		{big.NewInt(378221), big.NewInt(613), big.NewInt(617)},
		{big.NewInt(15), big.NewInt(3), big.NewInt(5)},

		// Testing vector taken from telegram docs.
		// * https://core.telegram.org/mtproto/samples-auth_key#4-encrypted-data-generation
		{big.NewInt(0x17ED48941A08F981), big.NewInt(0x494C553B), big.NewInt(0x53911073)},
	} {
		rnd := rand.New(rand.NewSource(239))
		p, q, err := DecomposePQ(tt.pq, rnd)
		if err != nil {
			t.Fatal(err)
		}
		if tt.p.Cmp(p) != 0 || tt.q.Cmp(q) != 0 {
			t.Errorf("PQ mismatch: %v %v, want %v %v", p, q, tt.p, tt.q)
		}
	}
}

func BenchmarkDecomposePQ(b *testing.B) {
	// DecomposePQ is used as Proof of Work so it is not required to be
	// very fast, leaving this benchmark here as a reference.
	//
	// This runs at ~300ms on Intel 8700k and allocates ~30mb, probably we
	// can at least reduce allocations.
	b.ReportAllocs()

	pq := big.NewInt(0x17ED48941A08F981)
	rnd := rand.New(rand.NewSource(239))
	for i := 0; i < b.N; i++ {
		p, q, err := DecomposePQ(pq, rnd)
		if err != nil {
			b.Fatal(err)
		}
		if p == nil || q == nil {
			b.Fatal("nil")
		}
	}
}
