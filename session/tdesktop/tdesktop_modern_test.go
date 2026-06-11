package tdesktop

import (
	"bytes"
	"encoding/binary"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/crypto"
)

// buildModernKeyData builds the body of a modern "key_data" file (the part
// after the TDF$ magic/version header), containing a single account.
//
// See https://github.com/telegramdesktop/tdesktop/blob/dev/Telegram/SourceFiles/storage/storage_domain.cpp.
func buildModernKeyData(t *testing.T, passcode, salt []byte, localKey crypto.Key) []byte {
	t.Helper()
	a := require.New(t)

	passcodeKey := createLocalKey(passcode, salt)

	// keyInnerData is a little-endian array whose first 256 bytes are the local key.
	var keyInner bytes.Buffer
	data := make([]byte, 272-4) // padded so the encrypted blob is a multiple of 16
	copy(data, localKey[:])
	a.NoError(writeArray(&keyInner, data, binary.LittleEndian))
	keyEncrypted, err := encryptLocal(keyInner.Bytes(), passcodeKey)
	a.NoError(err)

	// info: [length][count][index...]; one account with index 0.
	info := []byte{
		16, 0, 0, 0, // length (skipped on read)
		0, 0, 0, 1, // count = 1
		0, 0, 0, 0, // index = 0
		0, 0, 0, 0, // padding to 16 bytes
	}
	infoEncrypted, err := encryptLocal(info, localKey)
	a.NoError(err)

	var body bytes.Buffer
	a.NoError(writeArray(&body, salt, binary.BigEndian))
	a.NoError(writeArray(&body, keyEncrypted, binary.BigEndian))
	a.NoError(writeArray(&body, infoEncrypted, binary.BigEndian))
	return body.Bytes()
}

// buildModernMTPData builds the body of an account MTP data file (the part
// after the TDF$ magic/version header) holding a single auth key.
//
// See https://github.com/telegramdesktop/tdesktop/blob/v2.9.8/Telegram/SourceFiles/storage/storage_account.cpp.
func buildModernMTPData(t *testing.T, localKey crypto.Key, userID uint64, mainDC int, authKey crypto.Key) []byte {
	t.Helper()
	a := require.New(t)

	// Qt streams are big-endian; this mirrors qtReader in dbi.go.
	var auth bytes.Buffer
	put32 := func(v uint32) {
		a.NoError(binary.Write(&auth, binary.BigEndian, v))
	}
	put32(dbiMtpAuthorization)
	put32(0)               // mainLength, skipped on read
	put32(uint32(userID))  // legacyUserID (not the wide-ids tag, so used directly)
	put32(uint32(mainDC))  // legacyMainDCID
	put32(1)               // number of keys
	put32(uint32(mainDC))  // DC ID
	auth.Write(authKey[:]) // 256-byte auth key

	// Inner plaintext is a 4-byte length prefix (skipped) followed by the data,
	// padded to a multiple of the AES block size for encryptLocal.
	var inner bytes.Buffer
	a.NoError(binary.Write(&inner, binary.BigEndian, uint32(auth.Len())))
	inner.Write(auth.Bytes())
	for inner.Len()%16 != 0 {
		inner.WriteByte(0)
	}

	encrypted, err := encryptLocal(inner.Bytes(), localKey)
	a.NoError(err)

	var body bytes.Buffer
	a.NoError(writeArray(&body, encrypted, binary.BigEndian))
	return body.Bytes()
}

func wrapTDF(t *testing.T, body []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, writeFile(&buf, body, [4]byte{1, 0, 0, 0}))
	return buf.Bytes()
}

// TestReadFS_Modern verifies that the modern Telegram Desktop layout, where the
// key file is named "key_datas" (suffix 's') and there is no plain "key_data",
// is read correctly.
//
// See https://github.com/gotd/td/issues/825.
func TestReadFS_Modern(t *testing.T) {
	a := require.New(t)

	var passcode []byte
	salt := bytes.Repeat([]byte{0x11}, localEncryptSaltSize)

	var localKey crypto.Key
	for i := range localKey {
		localKey[i] = byte(i)
	}
	var authKey crypto.Key
	for i := range authKey {
		authKey[i] = byte(255 - i)
	}
	const (
		userID = 123456
		mainDC = 2
	)

	keyData := wrapTDF(t, buildModernKeyData(t, passcode, salt, localKey))
	mtpData := wrapTDF(t, buildModernMTPData(t, localKey, userID, mainDC, authKey))

	root := fstest.MapFS{
		// Modern layout: only the 's'-suffixed files exist, no plain "key_data".
		"key_datas":         {Data: keyData},
		"D877F783D5D3EF8Cs": {Data: mtpData}, // fileKey("data") + 's'
	}
	a.Equal("D877F783D5D3EF8C", fileKey("data"), "account file key must match tdesktop layout")

	accounts, err := ReadFS(root, passcode)
	a.NoError(err)
	a.Len(accounts, 1)

	got := accounts[0].Authorization
	a.Equal(uint64(userID), got.UserID)
	a.Equal(mainDC, got.MainDC)
	a.Equal(authKey, got.Keys[mainDC])
}

// TestReadFS_ModernPrecedence verifies that, when both the modern 's' file and a
// stale legacy '0' file are present, the modern one wins, matching Telegram
// Desktop's read order.
//
// See https://github.com/gotd/td/issues/825.
func TestReadFS_ModernPrecedence(t *testing.T) {
	a := require.New(t)

	var passcode []byte
	salt := bytes.Repeat([]byte{0x22}, localEncryptSaltSize)

	var localKey crypto.Key
	for i := range localKey {
		localKey[i] = byte(i)
	}
	var authKey crypto.Key
	for i := range authKey {
		authKey[i] = byte(255 - i)
	}
	const mainDC = 2

	keyData := wrapTDF(t, buildModernKeyData(t, passcode, salt, localKey))
	// Current (modern) and stale (legacy-suffixed) account files differ by UserID.
	current := wrapTDF(t, buildModernMTPData(t, localKey, 123, mainDC, authKey))
	stale := wrapTDF(t, buildModernMTPData(t, localKey, 999, mainDC, authKey))

	root := fstest.MapFS{
		"key_datas":         {Data: keyData},
		"D877F783D5D3EF8Cs": {Data: current}, // modern, must win
		"D877F783D5D3EF8C0": {Data: stale},   // stale legacy leftover
	}

	accounts, err := ReadFS(root, passcode)
	a.NoError(err)
	a.Len(accounts, 1)
	a.Equal(uint64(123), accounts[0].Authorization.UserID, "modern 's' file must take precedence over legacy '0'")
}
