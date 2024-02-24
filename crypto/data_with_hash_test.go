package crypto

import (
	"bytes"
	mathrand "math/rand"
	"testing"
)

func TestGuessData(t *testing.T) {
	for _, s := range []string{
		"a",
		"foo",
		"bar",
		"wake up neo",
		"24145",
	} {
		rnd := mathrand.New(mathrand.NewSource(239))
		data := []byte(s)
		dataWithHash, err := DataWithHash(data, rnd)
		if err != nil {
			t.Fatal(err)
		}
		guessed := GuessDataWithHash(dataWithHash)
		if guessed == nil {
			t.Fatal("got nil")
		}
		if !bytes.Equal(guessed, data) {
			t.Fatal("invalid guess")
		}
	}
}
