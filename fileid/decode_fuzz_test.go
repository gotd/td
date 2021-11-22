package fileid

import (
	"testing"
)

func FuzzDecodeFileID(f *testing.F) {
	for _, input := range testData {
		f.Add(input)
	}

	f.Fuzz(func(t *testing.T, fileID string) {
		_, err := DecodeFileID(fileID)
		if err != nil {
			t.Skip(err)
		}
	})
}
