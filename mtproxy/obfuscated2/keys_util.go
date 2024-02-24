package obfuscated2

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"io"
)

func createCTR(key, iv []byte) (cipher.Stream, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewCTR(block, iv), nil
}

func getDecryptInit(init []byte) (initRev [48]byte) {
	copy(initRev[:], init[8:56])
	// https://github.com/golang/go/wiki/SliceTricks#reversing
	for left, right := 0, len(initRev)-1; left < right; left, right = left+1, right-1 {
		initRev[left], initRev[right] = initRev[right], initRev[left]
	}

	return
}

// function from https://core.telegram.org/mtproto/mtproto-transports#transport-obfuscation
func generateInit(randSource io.Reader) (init [64]byte, err error) {
	// init := (56 random bytes) + protocol + dc + (2 random bytes)
	for {
		_, err = io.ReadFull(randSource, init[:])
		if err != nil {
			return [64]byte{}, err
		}

		// Filter some start sequences
		// See https://github.com/DrKLO/Telegram/blob/master/TMessagesProj/jni/tgnet/Connection.cpp#L531.
		// See https://github.com/tdlib/td/blob/master/td/mtproto/TcpTransport.cpp#L157-L158.
		if init[0] == 0xef { // Abridged header
			continue
		}

		firstInt := binary.LittleEndian.Uint32(init[0:4])
		if firstInt == 0x44414548 || // HEAD
			firstInt == 0x54534f50 || // POST
			firstInt == 0x20544547 || // GET
			firstInt == 0x4954504f || // OPTI
			firstInt == 0x02010316 || // ????
			firstInt == 0xdddddddd || // PaddedIntermediate header
			firstInt == 0xeeeeeeee /* Intermediate header */ {
			continue
		}

		if secondInt := binary.LittleEndian.Uint32(init[4:8]); secondInt == 0 {
			continue
		}

		break
	}

	return init, nil
}
