package tdesktop

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
)

var (
	testIP         = "127.0.0.1"
	optionTestData = func() []byte {
		testData := []uint8{
			0x00, 0x00, 0x00, 0x02, // ID
			0x00, 0x00, 0x00, 1 << 4, // Flags
			0x00, 0x00, 0x00, 80, // Port
			0x00, 0x00, 0x00, uint8(len(testIP)), // IP size
		}

		// IP
		testData = append(testData, testIP...)
		// Secret length
		testData = append(testData, 0x00, 0x00, 0x00, 0x00)

		return testData
	}()
)

func TestMTPDCOption_deserialize(t *testing.T) {
	maxCut := len(optionTestData)

	t.Run("OK", func(t *testing.T) {
		a := require.New(t)
		var m MTPDCOption
		a.NoError(m.deserialize(&qtReader{buf: bin.Buffer{Buf: optionTestData}}, 1))
		a.Equal(int32(2), m.ID)
		a.True(m.Static())
		a.Equal(bin.Fields(1<<4), m.Flags)
		a.Equal(int32(80), m.Port)
		a.Equal(testIP, m.IP)
	})

	for i := 0; i < maxCut; i += 4 {
		t.Run(fmt.Sprintf("EOFAfter%d", i), func(t *testing.T) {
			a := require.New(t)
			var m MTPDCOption
			r := &qtReader{buf: bin.Buffer{Buf: optionTestData[:i]}}
			a.Error(m.deserialize(r, 1))
		})
	}
}
