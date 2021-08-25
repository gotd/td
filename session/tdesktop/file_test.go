package tdesktop

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

var (
	testData = bytes.Repeat([]byte("abcd"), 4)
	zeros    = make([]byte, 32)
)

func Test_open(t *testing.T) {
	b := bytes.NewBuffer(nil)
	if err := writeFile(b, testData, [4]byte{}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		fs       fs.FS
		fileName string
		wantErr  bool
	}{
		{"Empty", fstest.MapFS{}, "testfile", true},
		{"SkipInvalidLength", fstest.MapFS{
			// open should skip first file and read next
			"testfile0": &fstest.MapFile{
				Data: zeros[:4],
			},
			"testfile1": &fstest.MapFile{
				Data: b.Bytes(),
			},
		}, "testfile", false},
		{"SkipInvalidMagic", fstest.MapFS{
			// open should skip first file and read next
			"testfile0": &fstest.MapFile{
				Data: zeros,
			},
			"testfile1": &fstest.MapFile{
				Data: b.Bytes(),
			},
		}, "testfile", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)
			if _, err := open(tt.fs, tt.fileName); tt.wantErr {
				a.Error(err)
			} else {
				a.NoError(err)
			}
		})
	}
}

func Test_fromFile(t *testing.T) {
	justBytes := func(data []byte, datas ...[]byte) func() io.Reader {
		return func() (r io.Reader) {
			r = bytes.NewReader(data)
			for _, d := range datas {
				r = io.MultiReader(r, bytes.NewReader(d))
			}
			return r
		}
	}

	tests := []struct {
		name    string
		data    func() io.Reader
		wantErr bool
	}{
		{"Empty", justBytes(nil), true},
		{"InvalidMagic", justBytes(zeros), true},
		{"InvalidLength", justBytes(tdesktopFileMagic[:], zeros[:8]), true},
		{"InvalidHash", justBytes(tdesktopFileMagic[:], zeros), true},
		{"WriteFile", func() io.Reader {
			b := bytes.NewBuffer(nil)
			if err := writeFile(b, testData, [4]byte{}); err != nil {
				t.Fatal(err)
			}
			return b
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)
			if r, err := fromFile(tt.data()); tt.wantErr {
				a.Error(err)
			} else {
				a.NoError(err)

				data, err := io.ReadAll(r)
				a.NoError(err)
				a.Equal(testData, data)
			}
		})
	}
}

func Test_telegramFileHash(t *testing.T) {
	data := bytes.Repeat([]byte{'a'}, 100)
	expected := [16]uint8{
		0xa8, 0xa9, 0xa1, 0x38,
		0xcf, 0x32, 0x37, 0xa9,
		0x4b, 0x78, 0xd0, 0x2d,
		0x03, 0xe0, 0x16, 0x81,
	}
	require.Equal(t, expected, telegramFileHash(data, [4]byte{0, 0, 0, 1}))
}

func Test_readArray(t *testing.T) {
	var validData bytes.Buffer
	if err := writeArray(&validData, testData, binary.LittleEndian); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		data    []byte
		expect  []byte
		order   binary.ByteOrder
		wantErr bool
	}{
		{"LengthEOF", nil, nil, binary.LittleEndian, true},
		{"DataEOF", []byte{255, 0, 0, 0}, nil, binary.LittleEndian, true},
		{"OK", validData.Bytes(), testData, binary.LittleEndian, false},
		{"0xffffffff", []byte{0xff, 0xff, 0xff, 0xff}, nil, binary.LittleEndian, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)
			if data, err := readArray(bytes.NewReader(tt.data), tt.order); tt.wantErr {
				a.Error(err)
			} else {
				a.NoError(err)
				a.Equal(tt.expect, data)
			}
		})
	}
}
