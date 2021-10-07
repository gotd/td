package tdesktop

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
)

func Test_mtpAuthorization_deserialize(t *testing.T) {
	testData := []uint8{
		0x00, 0x00, 0x00, 0x4b, // dbiMtpAuthorization
		0x00, 0x00, 0x05, 0x30, // mainLength
		0xff, 0xff, 0xff, 0xff, // legacyUserId
		0xff, 0xff, 0xff, 0xff, // legacyMainDcId
		0x00, 0x00, 0x00, 0x00, 0x12, 0x73, 0xab, 0x45, // UserID = 309570373
		0x00, 0x00, 0x00, 0x02, // DC = 2
		0x00, 0x00, 0x00, 0x05, // 5 keys.
	}
	maxCut := len(testData) + 16
	for i := byte(0); i < 5; i++ {
		testData = append(testData, 0, 0, 0, i) // DC ID as BigEndian uint32
		key := bytes.Repeat([]byte{i}, 256)
		testData = append(testData, key...)
	}

	t.Run("OK", func(t *testing.T) {
		a := require.New(t)
		var m MTPAuthorization
		a.NoError(m.deserialize(&reader{buf: bin.Buffer{Buf: testData}}))
		a.Equal(int64(309570373), m.UserID)
		a.Equal(2, m.MainDC)
		a.Len(m.Keys, 5)
		for i := 0; i < 5; i++ {
			a.Equal(m.Keys[i][0], uint8(i))
		}
	})
	t.Run("WrongID", func(t *testing.T) {
		a := require.New(t)
		var m MTPAuthorization
		a.Error(m.deserialize(&reader{buf: bin.Buffer{Buf: make([]byte, 4)}}))
	})

	for i := 0; i < maxCut; i += 4 {
		t.Run(fmt.Sprintf("EOFAfter%d", i), func(t *testing.T) {
			a := require.New(t)
			var m MTPAuthorization
			a.Error(m.deserialize(&reader{buf: bin.Buffer{Buf: testData[:i]}}))
		})
	}
}
