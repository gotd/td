package tdesktop

import (
	"bytes"
	"crypto/aes"
	"crypto/sha1" // #nosec G505
	"crypto/sha512"

	"github.com/go-faster/errors"
	"golang.org/x/crypto/pbkdf2"

	"github.com/gotd/ige"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
)

// See https://github.com/telegramdesktop/tdesktop/blob/v2.9.8/Telegram/SourceFiles/storage/details/storage_file_utilities.cpp#L322.
func createLegacyLocalKey(passcode, salt []byte) (r crypto.Key) {
	iters := localEncryptNoPwdIterCount
	if len(passcode) > 0 {
		iters = localEncryptIterCount
	}

	key := pbkdf2.Key(passcode, salt, iters, len(r), sha1.New)
	copy(r[:], key)
	return r
}

// See https://github.com/telegramdesktop/tdesktop/blob/v2.9.8/Telegram/SourceFiles/storage/details/storage_file_utilities.cpp#L300.
func createLocalKey(passcode, salt []byte) (r crypto.Key) {
	iters := 1
	if len(passcode) > 0 {
		iters = kStrongIterationsCount
	}

	h := sha512.New()
	_, _ = h.Write(salt)
	_, _ = h.Write(passcode)
	_, _ = h.Write(salt)

	key := pbkdf2.Key(h.Sum(nil), salt, iters, len(r), sha512.New)
	copy(r[:], key)
	return r
}

// See https://github.com/telegramdesktop/tdesktop/blob/v2.9.8/Telegram/SourceFiles/storage/details/storage_file_utilities.cpp#L584.
func decryptLocal(encrypted []byte, localKey crypto.Key) ([]byte, error) {
	if l := len(encrypted); l%aes.BlockSize != 0 {
		return nil, errors.Errorf("invalid length %d, must be padded to 16", l)
	}
	// Get encryptedKey.
	var msgKey bin.Int128
	n := copy(msgKey[:], encrypted)
	encrypted = encrypted[n:]

	aesKey, aesIV := crypto.OldKeys(localKey, msgKey, crypto.Server)
	cipher, err := aes.NewCipher(aesKey[:])
	if err != nil {
		return nil, errors.Wrap(err, "create cipher")
	}

	decrypted := make([]byte, len(encrypted))
	ige.DecryptBlocks(cipher, aesIV[:], decrypted, encrypted)

	if h := sha1.Sum(decrypted); !bytes.Equal(h[:16], msgKey[:]) /* #nosec G401 */ {
		return nil, errors.New("msg_key mismatch")
	}
	return decrypted, nil
}

// encryptLocal code may panic
func encryptLocal(decrypted []byte, localKey crypto.Key) ([]byte, error) {
	if l := len(decrypted); l%aes.BlockSize != 0 {
		return nil, errors.Errorf("invalid length %d, must be padded to 16", l)
	}
	// Compute encryptedKey.
	var msgKey bin.Int128
	h := sha1.Sum(decrypted) // #nosec G401
	copy(msgKey[:], h[:])

	aesKey, aesIV := crypto.OldKeys(localKey, msgKey, crypto.Server)
	cipher, err := aes.NewCipher(aesKey[:])
	if err != nil {
		return nil, errors.Wrap(err, "create cipher")
	}

	encrypted := make([]byte, 16+len(decrypted))
	copy(encrypted, msgKey[:])
	ige.EncryptBlocks(cipher, aesIV[:], encrypted[16:], decrypted)

	return encrypted, nil
}
