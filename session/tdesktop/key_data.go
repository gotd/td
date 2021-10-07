package tdesktop

import (
	"bytes"
	"encoding/binary"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/crypto"
)

type keyData struct {
	localKey    crypto.Key
	accountsIDx []uint32
}

// See https://github.com/telegramdesktop/tdesktop/blob/v2.9.8/Telegram/SourceFiles/storage/storage_domain.cpp#L119-L159.
func readKeyData(tgf tdesktopFile, passcode []byte) (_ keyData, rErr error) {
	salt, err := tgf.readArray()
	if err != nil {
		return keyData{}, xerrors.Errorf("read salt: %w", err)
	}
	if l := len(salt); l != localEncryptSaltSize {
		return keyData{}, xerrors.Errorf("invalid salt length %d", l)
	}

	passcodeKey := createLocalKey(passcode, salt)
	keyEncrypted, err := tgf.readArray()
	if err != nil {
		return keyData{}, xerrors.Errorf("read keyEncrypted: %w", err)
	}
	keyInnerData, err := decryptLocal(keyEncrypted, passcodeKey)
	if err != nil {
		return keyData{}, xerrors.Errorf("decrypt keyEncrypted: %w", err)
	}
	key, err := readArray(bytes.NewReader(keyInnerData), binary.LittleEndian)
	if err != nil {
		return keyData{}, xerrors.Errorf("read key: %w", err)
	}

	if l := len(key); l < len(crypto.Key{}) {
		return keyData{}, xerrors.Errorf("key too small (%d)", l)
	}
	var localKey crypto.Key
	copy(localKey[:], key)

	infoEncrypted, err := tgf.readArray()
	if err != nil {
		return keyData{}, xerrors.Errorf("read infoEncrypted: %w", err)
	}
	infoDecrypted, err := decryptLocal(infoEncrypted, localKey)
	if err != nil {
		return keyData{}, ErrKeyInfoDecrypt
	}
	// Skip decrypted data length.
	infoDecrypted = infoDecrypted[4:]
	// Read count of accounts.
	count := int(binary.BigEndian.Uint32(infoDecrypted))
	infoDecrypted = infoDecrypted[4:]

	// Preallocate accountsIDx.
	accountsIDx := make([]uint32, 0, count)
	for i := 0; i < count; i++ {
		idx := binary.BigEndian.Uint32(infoDecrypted[i*4:])
		accountsIDx = append(accountsIDx, idx)
	}

	return keyData{
		localKey:    localKey,
		accountsIDx: accountsIDx,
	}, nil
}
