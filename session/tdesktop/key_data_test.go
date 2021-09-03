package tdesktop

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_readKeyData(t *testing.T) {
	a := require.New(t)
	buf := bytes.Buffer{}
	var (
		passcode []byte
		salt     = make([]byte, 32)
		info     = []byte{
			16, 0, 0, 0,
			0, 0, 0, 1,
			0, 0, 0, 0,
			0, 0, 0, 0,
		}
		passcodeKey = createLocalKey(passcode, salt)
		localKey    = createLocalKey([]byte("aboba"), salt)

		// Store buffer offsets to test EOF errors
		cuts = []int{0}
	)

	// Write salt.
	a.NoError(writeArray(&buf, salt, binary.BigEndian))
	cuts = append(cuts, buf.Len())

	var keyInnerData []byte
	{
		b := bytes.Buffer{}
		// Pad to 16 bytes.
		data := make([]byte, 272-4)
		copy(data, localKey[:])
		a.NoError(writeArray(&b, data, binary.LittleEndian))
		keyInnerData = b.Bytes()
	}

	keyEncrypted, err := encryptLocal(keyInnerData, passcodeKey)
	a.NoError(err)
	a.NoError(writeArray(&buf, keyEncrypted, binary.BigEndian))
	cuts = append(cuts, buf.Len())

	infoEncrypted, err := encryptLocal(info, localKey)
	a.NoError(err)
	a.NoError(writeArray(&buf, infoEncrypted, binary.BigEndian))
	// Do not store last cut, because it is valid.

	fileData := buf.Bytes()
	t.Run("OK", func(t *testing.T) {
		a := require.New(t)

		kdata, err := readKeyData(&tdesktopFile{
			data: fileData,
		}, passcode)
		a.NoError(err)
		a.Equal(localKey, kdata.localKey)
		a.Len(kdata.accountsIDx, 1)
	})
	for _, cut := range cuts {
		t.Run(fmt.Sprintf("EOFAfter%d", cut), func(t *testing.T) {
			a := require.New(t)

			_, err := readKeyData(&tdesktopFile{
				data: fileData[:cut],
			}, passcode)
			a.Error(err)
		})
	}
}
