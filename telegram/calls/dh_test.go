package calls

import (
	"crypto/rand"
	"math/big"
	"testing"
)

// rfc3526Group2048 is the well-known RFC 3526 2048-bit MODP safe prime with
// generator 2. It is a stand-in for the Telegram DH group in unit tests.
const rfc3526Prime2048 = "FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD1" +
	"29024E088A67CC74020BBEA63B139B22514A08798E3404DD" +
	"EF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245" +
	"E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7ED" +
	"EE386BFB5A899FA5AE9F24117C4B1FE649286651ECE45B3D" +
	"C2007CB8A163BF0598DA48361C55D39A69163FA8FD24CF5F" +
	"83655D23DCA3AD961C62F356208552BB9ED529077096966D" +
	"670C354E4ABC9804F1746C08CA18217C32905E462E36CE3B" +
	"E39E772C180E86039B2783A2EC07A28FB5C55DF06F4C52C9" +
	"DE2BCBF6955817183995497CEA956AE515D2261898FA0510" +
	"15728E5A8AACAA68FFFFFFFFFFFFFFFF"

func testGroup(t *testing.T) *dhConfig {
	t.Helper()
	p, ok := new(big.Int).SetString(rfc3526Prime2048, 16)
	if !ok {
		t.Fatal("parse test prime")
	}
	return &dhConfig{g: 2, p: p}
}

func TestDHKeyAgreement(t *testing.T) {
	dh := testGroup(t)

	a, gAInt, err := dh.randomExp(rand.Reader, nil)
	if err != nil {
		t.Fatalf("caller exp: %v", err)
	}
	b, gBInt, err := dh.randomExp(rand.Reader, nil)
	if err != nil {
		t.Fatalf("callee exp: %v", err)
	}
	gA, gB := pad(gAInt), pad(gBInt)

	callerKey, callerFP, err := dh.computeKey(gB, a)
	if err != nil {
		t.Fatalf("caller key: %v", err)
	}
	calleeKey, calleeFP, err := dh.computeKey(gA, b)
	if err != nil {
		t.Fatalf("callee key: %v", err)
	}

	if len(callerKey) != keySize {
		t.Fatalf("key size = %d, want %d", len(callerKey), keySize)
	}
	if string(callerKey) != string(calleeKey) {
		t.Fatal("shared keys differ")
	}
	if callerFP != calleeFP {
		t.Fatalf("fingerprints differ: %d vs %d", callerFP, calleeFP)
	}
}

func TestDHRandomExpRange(t *testing.T) {
	dh := testGroup(t)
	for range 32 {
		exp, pub, err := dh.randomExp(rand.Reader, nil)
		if err != nil {
			t.Fatal(err)
		}
		if exp.Cmp(bigTwo) < 0 {
			t.Fatal("exponent below 2")
		}
		if err := dh.checkValue(pub); err != nil {
			t.Fatalf("public value out of range: %v", err)
		}
	}
}

func TestDHCheckValueRejectsConfinement(t *testing.T) {
	dh := testGroup(t)
	pMinusOne := new(big.Int).Sub(dh.p, bigOne)
	for _, v := range []*big.Int{
		big.NewInt(0), big.NewInt(1), pMinusOne, new(big.Int).Set(dh.p),
	} {
		if err := dh.checkValue(v); err == nil {
			t.Fatalf("checkValue accepted forbidden value %s", v)
		}
	}
}

func TestDHRandomExpMixesServerRandom(t *testing.T) {
	dh := testGroup(t)
	// With identical client randomness, differing server randomness must
	// produce different exponents.
	server := make([]byte, keySize)
	server[0] = 0xAA
	exp1, _, err := dh.randomExp(zeroReader{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	exp2, _, err := dh.randomExp(zeroReader{}, server)
	if err != nil {
		t.Fatal(err)
	}
	if exp1.Cmp(exp2) == 0 {
		t.Fatal("server random not mixed into exponent")
	}
}

// zeroReader yields an endless stream of zero bytes.
type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
