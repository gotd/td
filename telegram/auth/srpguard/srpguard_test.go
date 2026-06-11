package srpguard_test

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/auth/srpguard"
	"github.com/gotd/td/tg"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	require.NoError(t, err)
	return b
}

func testPassword(t *testing.T) *tg.AccountPassword {
	t.Helper()
	algo := &tg.PasswordKdfAlgoSHA256SHA256PBKDF2HMACSHA512iter100000SHA256ModPow{
		Salt1: mustHex(t, "4D11FB6BEC38F9D2546BB0F61E4F1C99A1BC0DB8F0D5F35B1291B37B213123D7ED48F3C6794D495B"),
		Salt2: mustHex(t, "A1B181AAFE88188680AE32860D60BB01"),
		G:     3,
		P: mustHex(t, "C71CAEB9C6B1C9048E6C522F70F13F73980D40238E3E21C14934D037563D930F"+
			"48198A0AA7C14058229493D22530F4DBFA336F6E0AC925139543AED44CCE7C37"+
			"20FD51F69458705AC68CD4FE6B6B13ABDC9746512969328454F18FAF8C595F64"+
			"2477FE96BB2A941D5BCD1D4AC8CC49880708FA9B378E3C4F3A9060BEE67CF9A4"+
			"A4A695811051907E162753B56B0F6B410DBA74D8A84B2A14B3144E0EF1284754"+
			"FD17ED950D5965B4B9DD46582DB1178D169C6BC465B0D6FF9CA3928FEF5B9AE4"+
			"E418FC15E83EBEA0F87FA9FF5EED70050DED2849F47BF959D956850CE929851F"+
			"0D8115F635B105EE2E4E15D04B2454BF6F4FADF034B10403119CD8E3B92FCC5B"),
	}
	p := &tg.AccountPassword{SRPID: 1234567890}
	p.SetCurrentAlgo(algo)
	return p
}

func TestLockedBuffer(t *testing.T) {
	p := testPassword(t)
	buf := memguard.NewBufferFromBytes([]byte("correct horse battery staple"))

	answer, err := srpguard.LockedBuffer(buf)(context.Background(), p)
	require.NoError(t, err)
	require.NotEmpty(t, answer.A)
	require.NotEmpty(t, answer.M1)
	require.Equal(t, p.SRPID, answer.SRPID)
	require.False(t, buf.IsAlive(), "buffer must be destroyed after use")
}

func TestEnclave(t *testing.T) {
	p := testPassword(t)
	enc := memguard.NewEnclave([]byte("correct horse battery staple"))

	answer, err := srpguard.Enclave(enc)(context.Background(), p)
	require.NoError(t, err)
	require.NotEmpty(t, answer.A)
	require.NotEmpty(t, answer.M1)

	// The enclave may be reused for a second attempt.
	answer2, err := srpguard.Enclave(enc)(context.Background(), p)
	require.NoError(t, err)
	require.NotEmpty(t, answer2.A)
}

func TestLockedBufferDestroyed(t *testing.T) {
	buf := memguard.NewBufferFromBytes([]byte("x"))
	buf.Destroy()

	_, err := srpguard.LockedBuffer(buf)(context.Background(), testPassword(t))
	require.Error(t, err)
}
