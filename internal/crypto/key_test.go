package crypto

import (
	"testing"

	"github.com/nnqq/td/bin"
)

func TestAuthKeyID(t *testing.T) {
	var k Key
	for i := 0; i < 256; i++ {
		k[i] = byte(i)
	}

	if k.ID() != [8]byte{50, 209, 88, 110, 164, 87, 223, 200} {
		t.Error("bad id")
	}
	if k.AuxHash() != [8]byte{73, 22, 214, 189, 183, 247, 142, 104} {
		t.Error("bad aux hash")
	}
}

func TestCalcKey(t *testing.T) {
	var k Key
	for i := 0; i < 256; i++ {
		k[i] = byte(i)
	}
	var m bin.Int128
	for i := 0; i < 16; i++ {
		m[i] = byte(i)
	}

	t.Run("Client", func(t *testing.T) {
		key, iv := Keys(k, m, Client)
		if key != [32]byte{
			112, 78, 208, 156, 139, 65, 102, 138, 232, 249, 157, 36, 71, 56, 247, 29,
			189, 220, 68, 70, 155, 107, 189, 74, 168, 87, 61, 208, 66, 189, 5, 158,
		} {
			t.Error("bad key")
		}
		if iv != [32]byte{
			77, 38, 96, 0, 165, 80, 237, 171, 191, 76, 124, 228, 15, 208, 4, 60, 201, 34, 48,
			24, 76, 211, 23, 165, 204, 156, 36, 130, 253, 59, 147, 24,
		} {
			t.Error("bad iv")
		}
	})
	t.Run("Server", func(t *testing.T) {
		key, iv := Keys(k, m, Server)
		if key != [32]byte{
			33, 119, 37, 121, 155, 36, 88, 6, 69, 129, 116, 161, 252, 251, 200, 131, 144, 104,
			7, 177, 80, 51, 253, 208, 234, 43, 77, 105, 207, 156, 54, 78,
		} {
			t.Error("bad key")
		}
		if iv != [32]byte{
			102, 154, 101, 56, 145, 122, 79, 165, 108, 163, 35, 96, 164, 49, 201, 22, 11, 228,
			173, 136, 113, 64, 152, 13, 171, 145, 206, 123, 220, 71, 255, 188,
		} {
			t.Error("bad iv")
		}
	})
}
