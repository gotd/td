package obfuscated2

import (
	"crypto/cipher"
	"encoding/binary"
	"io"

	"github.com/go-faster/errors"

	"github.com/gotd/td/crypto"
)

type keys struct {
	header  []byte
	encrypt cipher.Stream
	decrypt cipher.Stream
}

func (k *keys) createStreams(init, secret []byte) error {
	// preallocate 256 bit key + 16 byte secret
	const keyLength = 32 + 16

	encryptKey := append(make([]byte, 0, keyLength), init[8:40]...)
	encryptIV := append(make([]byte, 0, 16), init[40:56]...)

	initRev := getDecryptInit(init)
	decryptKey := append(make([]byte, 0, keyLength), initRev[:32]...)
	decryptIV := append(make([]byte, 0, 16), initRev[32:48]...)

	if len(secret) > 0 {
		if len(secret) < 16 {
			return errors.Errorf("invalid secret size %d", len(secret))
		}
		secret = secret[0:16]

		encryptKey = crypto.SHA256(encryptKey, secret)
		decryptKey = crypto.SHA256(decryptKey, secret)
	}

	var err error
	k.encrypt, err = createCTR(encryptKey, encryptIV)
	if err != nil {
		return err
	}

	k.decrypt, err = createCTR(decryptKey, decryptIV)
	if err != nil {
		return err
	}

	return nil
}

func generateKeys(randSource io.Reader, protocol [4]byte, secret []byte, dc int) (keys, error) {
	init, err := generateInit(randSource)
	if err != nil {
		return keys{}, err
	}

	var k keys
	if err := k.createStreams(init[:], secret); err != nil {
		return keys{}, err
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
