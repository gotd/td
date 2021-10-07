package crypto

import (
	"math/big"
	"testing"

	"github.com/nnqq/td/bin"
)

func TestEncryptedMessage_Encode(t *testing.T) {
	k, ok := big.NewInt(0).SetString(`644475571b8fac33f5072049f29d3eeb4493cea84e925d0601c31c1edbb79567adf23c7b97f7882d70f23cff5b8d62eff66399cd32f35b1882ac602e76f30701975c73ad70937169d840b9483e306ab49e656826b2aedc4451d20d65fe96120ecd97ccc16e6ef8ce12cb90c37db21f9c1700ee282f2fba088af1491a3b7d93a2f7abb496e5015779d8c107c2a61d8f992c909b52d29be44ac55d4d077351c96591bfa44a3482d90080ad4bd1417300c88c715f28b03c7b7f1e6ddffd0f321df64adcfdf6f99c756f2df8a7bf9f55110b7353342e050ffb1353afc9a888d10a0287b7a5d94368ba2eb6f39730745905ce42c63d3950e97acd190bd20cc030182e`, 16)
	if !ok {
		t.Fatal(ok)
	}

	payload := []byte{1, 2, 3, 4}

	var authKey Key
	k.FillBytes(authKey[:])

	d := EncryptedMessage{
		EncryptedData: payload,
		MsgKey:        bin.Int128{0, 0, 0, 0},
		AuthKeyID:     authKey.ID(),
	}
	b := new(bin.Buffer)
	if err := d.Encode(b); err != nil {
		t.Fatal(err)
	}
}
