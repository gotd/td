package obfuscated2

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"io"

	"github.com/gotd/td/internal/crypto"
)

func createCTR(key, iv []byte) (stream cipher.Stream, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	stream = cipher.NewCTR(block, iv)
	return
}

func getDecryptInit(init [64]byte) (initRev [48]byte) {
	copy(initRev[:], init[8:56])
	// https://github.com/golang/go/wiki/SliceTricks#reversing
	for left, right := 0, len(initRev)-1; left < right; left, right = left+1, right-1 {
		initRev[left], initRev[right] = initRev[right], initRev[left]
	}

	return
}

type keys struct {
	header  []byte
	encrypt cipher.Stream
	decrypt cipher.Stream
}

func generateKeys(randSource io.Reader, protocol [4]byte, secret []byte, dc int) (k keys, err error) {
	init, err := generateInit(randSource)
	if err != nil {
		return
	}

	// preallocate 256 bit key + 16 bit secret
	const keyLength = 32 + 16

	encryptKey := append(make([]byte, 0, keyLength), init[8:40]...)
	encryptIV := append(make([]byte, 0, 16), init[40:56]...)

	initRev := getDecryptInit(init)
	decryptKey := append(make([]byte, 0, keyLength), initRev[:32]...)
	decryptIV := append(make([]byte, 0, 16), initRev[32:48]...)
	secret = secret[0:16]

	encryptKey = crypto.SHA256(encryptKey, secret)
	decryptKey = crypto.SHA256(decryptKey, secret)

	k.encrypt, err = createCTR(encryptKey, encryptIV)
	if err != nil {
		return
	}

	k.decrypt, err = createCTR(decryptKey, decryptIV)
	if err != nil {
		return
	}

	copy(init[56:60], protocol[:])
	binary.LittleEndian.PutUint16(init[60:62], uint16(dc))

	var encryptedInit [64]byte
	k.encrypt.XORKeyStream(encryptedInit[:], init[:])
	k.header = make([]byte, 64)
	copy(k.header, init[0:56])
	copy(k.header[56:], encryptedInit[56:56+8])

	return k, nil
}

// function from https://core.telegram.org/mtproto/mtproto-transports#transport-obfuscation
func generateInit(randSource io.Reader) (init [64]byte, err error) {
	// init := (56 random bytes) + protocol + dc + (2 random bytes)
	for {
		_, err = io.ReadFull(randSource, init[:])
		if err != nil {
			return
		}

		if init[0] == 0xef {
			continue
		}

		firstInt := binary.LittleEndian.Uint32(init[0:4])
		if firstInt == 0x44414548 ||
			firstInt == 0x54534f50 ||
			firstInt == 0x20544547 ||
			firstInt == 0x4954504f ||
			firstInt == 0x02010316 ||
			firstInt == 0xdddddddd ||
			firstInt == 0xeeeeeeee {
			continue
		}

		if secondInt := binary.LittleEndian.Uint32(init[4:8]); secondInt == 0 {
			continue
		}

		break
	}

	return init, nil
}
