package crypto

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

// From Telegram docs example.
//
// See https://core.telegram.org/mtproto/auth_key#presenting-proof-of-work-server-authentication.
var (
	checkGPdhPrime = func() *big.Int {
		data, err := hex.DecodeString("C71CAEB9C6B1C9048E6C522F70F13F73980D40238E3E21C14934D037563D930F" +
			"48198A0AA7C14058229493D22530F4DBFA336F6E0AC925139543AED44CCE7C37" +
			"20FD51F69458705AC68CD4FE6B6B13ABDC9746512969328454F18FAF8C595F64" +
			"2477FE96BB2A941D5BCD1D4AC8CC49880708FA9B378E3C4F3A9060BEE67CF9A4" +
			"A4A695811051907E162753B56B0F6B410DBA74D8A84B2A14B3144E0EF1284754" +
			"FD17ED950D5965B4B9DD46582DB1178D169C6BC465B0D6FF9CA3928FEF5B9AE4" +
			"E418FC15E83EBEA0F87FA9FF5EED70050DED2849F47BF959D956850CE929851F" +
			"0D8115F635B105EE2E4E15D04B2454BF6F4FADF034B10403119CD8E3B92FCC5B")
		if err != nil {
			panic(err)
		}

		return big.NewInt(0).SetBytes(data)
	}()
	checkGPg = 3
)

func BenchmarkCheckGP(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = CheckGP(checkGPg, checkGPdhPrime)
	}
}

func TestCheckGP(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		a := require.New(t)
		a.NoError(CheckGP(checkGPg, checkGPdhPrime))
	})
	t.Run("WrongG", func(t *testing.T) {
		require.Error(t, CheckGP(1337, checkGPdhPrime))
	})
	t.Run("WrongDivider", func(t *testing.T) {
		// CheckGP should check that p mod 3 = 2 for g = 3;
		// We pass p = 4, so p mod 3 = 1.
		require.Error(t, CheckGP(3, big.NewInt(4)))
	})
}
