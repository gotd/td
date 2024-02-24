package srp

import (
	"math/big"

	"github.com/gotd/td/crypto"
)

func (s SRP) pad256FromBig(i *big.Int) (b [256]byte, r bool) {
	r = crypto.FillBytes(i, b[:])
	return b, r
}

func (s SRP) pad256(b []byte) [256]byte {
	if len(b) >= 256 {
		return *(*[256]byte)(b[len(b)-256:])
	}

	var tmp [256]byte
	copy(tmp[256-len(b):], b)

	return tmp
}
