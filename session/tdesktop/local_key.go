package tdesktop

import (
	"bytes"
	"crypto/aes"
	"crypto/sha1"
	"crypto/sha512"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/xerrors"

	"github.com/gotd/ige"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
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
	{
		_, _ = h.Write(salt)
		_, _ = h.Write(passcode)
		_, _ = h.Write(salt)
	}

	key := pbkdf2.Key(h.Sum(nil), salt, iters, len(r), sha512.New)
	copy(r[:], key)
	return r
}

// See https://github.com/telegramdesktop/tdesktop/blob/v2.9.8/Telegram/SourceFiles/storage/details/storage_file_utilities.cpp#L584.
func decryptLocal(encrypted []byte, localKey crypto.Key) ([]byte, error) {
	// Get encryptedKey.
	var msgKey bin.Int128
	n := copy(msgKey[:], encrypted)
	encrypted = encrypted[n:]

	aesKey, aesIV := crypto.OldKeys(localKey, msgKey, crypto.Server)
	cipher, err := aes.NewCipher(aesKey[:])
	if err != nil {
		return nil, xerrors.Errorf("create cipher: %w", err)
	}

	decrypted := make([]byte, len(encrypted))
	ige.DecryptBlocks(cipher, aesIV[:], decrypted, encrypted)

	if h := sha1.Sum(decrypted); !bytes.Equal(h[:16], msgKey[:]) {
		return nil, xerrors.New("msg_key mismatch")
	}
	return decrypted, nil
}
