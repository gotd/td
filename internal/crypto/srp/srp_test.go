package srp

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

type srpArgs struct {
	password []byte
	srpB     []byte
	mp       Input
}

func testSRPInput(t testing.TB) srpArgs {
	return srpArgs{
		password: []byte("123123"),
		srpB: getHex(t, "9C52401A6A8084EC82F01C3725D3FB448BD2F0C909F9D97726EAC4B7A74172D9"+
			"52F02466BE6734FA274D2B7429E27397F10372D66B400B80A5C5AE3F28B17BF3"+
			"105D7A2D2A885998CDC2DEFC208AEC217AB58859A9ABC2374AD93DC285F4B3FB"+
			"CAFF4143D7888F2425BD2FB711B25609CEB21757D935B1EF2F042173AD0CE2FE"+
			"0E474DAC53914BD25A8A9AED4AEA8953D55CB88621DB37B871EA0D04393AC098"+
			"7F68094CCC9DE8239251375D8FFFD263316CD528C097B7BC9FB919FBEDB76C52"+
			"5DF3413C374EE076D97A1E6D352BB7CC80FD13651B04B32E2E48C5268150842C"+
			"FD07CF855958B1B5EA9C36FDAD697FE3AEC8DCC6B1EFEC36874AF226204676CF"),
		mp: Input{
			Salt1: getHex(t, "4D11FB6BEC38F9D2546BB0F61E4F1C99A1BC0DB8F0D5F35B1291B37B213123D7ED48F3C6794D495B"),
			Salt2: getHex(t, "A1B181AAFE88188680AE32860D60BB01"),
			G:     3,
			P: getHex(t, "C71CAEB9C6B1C9048E6C522F70F13F73980D40238E3E21C14934D037563D930F"+
				"48198A0AA7C14058229493D22530F4DBFA336F6E0AC925139543AED44CCE7C37"+
				"20FD51F69458705AC68CD4FE6B6B13ABDC9746512969328454F18FAF8C595F64"+
				"2477FE96BB2A941D5BCD1D4AC8CC49880708FA9B378E3C4F3A9060BEE67CF9A4"+
				"A4A695811051907E162753B56B0F6B410DBA74D8A84B2A14B3144E0EF1284754"+
				"FD17ED950D5965B4B9DD46582DB1178D169C6BC465B0D6FF9CA3928FEF5B9AE4"+
				"E418FC15E83EBEA0F87FA9FF5EED70050DED2849F47BF959D956850CE929851F"+
				"0D8115F635B105EE2E4E15D04B2454BF6F4FADF034B10403119CD8E3B92FCC5B"),
		},
	}
}

func TestSRP(t *testing.T) {
	tests := []struct {
		args        srpArgs
		want        Answer
		expectError assert.ErrorAssertionFunc
	}{
		{
			args: testSRPInput(t),
			want: Answer{
				A:  setByte(256, 3),
				M1: getHex(t, "999DF906BDA2C6CBB52F503406EBA2D0D0503ACE0CC302C38F13EE5010AD4051"),
			},
			expectError: assert.NoError,
		},
	}
	for i := range tests {
		tcase := tests[i]
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			random := setByte(256, 1)
			srp := NewSRP(rand.Reader)
			got, err := srp.Hash(tcase.args.password, tcase.args.srpB, random, tcase.args.mp)
			if !tcase.expectError(t, err) {
				return
			}

			if !assert.Equal(t, tcase.want, got) {
				return
			}
		})
	}
}

func getHex(t testing.TB, in string) []byte {
	res, err := hex.DecodeString(in)
	if err != nil {
		t.Fatal("failed to get hex", err)
	}
	return res
}

func setByte(size, value int) []byte {
	res := make([]byte, size)
	binary.BigEndian.PutUint32(res[size-4:], uint32(value))
	return res
}

const pBase64 = "xxyuucaxyQSObFIvcPE_c5gNQCOOPiHBSTTQN1Y9kw9IGYoKp8FAWCKUk9IlMPTb-jNvbgrJJROVQ67UTM58NyD9UfaUWHBaxozU_" +
	"mtrE6vcl0ZRKWkyhFTxj6-MWV9kJHf-lrsqlB1bzR1KyMxJiAcI-ps3jjxPOpBgvuZ8-aSkppWBEFGQfhYnU7VrD2tBDbp02KhLKhSzFE4O8ShHVP0" +
	"X7ZUNWWW0ud1GWC2xF40WnGvEZbDW_5yjko_vW5rk5Bj8Feg-vqD4f6n_Xu1wBQ3tKEn0e_lZ2VaFDOkphR8NgRX2NbEF7i5OFdBLJFS_b0-" +
	"t8DSxBAMRnNjjuS_MWw=="

func Test_checkInput(t *testing.T) {
	p, err := base64.URLEncoding.DecodeString(pBase64)
	if err != nil {
		t.Fatal("no err expected", err)
	}

	err = checkInput(3, big.NewInt(0).SetBytes(p))
	if err != nil {
		t.Fatal("no err expected", err)
	}
}

func BenchmarkSRP_Auth(b *testing.B) {
	input := testSRPInput(b)
	srp := NewSRP(rand.Reader)
	random := setByte(256, 1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = srp.Hash(input.password, input.srpB, random, input.mp)
	}
}
