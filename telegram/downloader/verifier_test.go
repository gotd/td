package downloader

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

var hashRanges = [][]tg.FileHash{
	{
		tg.FileHash{Offset: 0, Limit: 131072},
		tg.FileHash{Offset: 131072, Limit: 131072},
		tg.FileHash{Offset: 262144, Limit: 131072},
		tg.FileHash{Offset: 393216, Limit: 131072},
		tg.FileHash{Offset: 524288, Limit: 131072},
		tg.FileHash{Offset: 655360, Limit: 131072},
		tg.FileHash{Offset: 786432, Limit: 131072},
		tg.FileHash{Offset: 917504, Limit: 131072},
	}, {
		tg.FileHash{Offset: 1048576, Limit: 131072},
		tg.FileHash{Offset: 1179648, Limit: 131072},
		tg.FileHash{Offset: 1310720, Limit: 131072},
		tg.FileHash{Offset: 1441792, Limit: 131072},
		tg.FileHash{Offset: 1572864, Limit: 131072},
		tg.FileHash{Offset: 1703936, Limit: 131072},
		tg.FileHash{Offset: 1835008, Limit: 131072},
		tg.FileHash{Offset: 1966080, Limit: 131072},
	}, {
		tg.FileHash{Offset: 2097152, Limit: 131072},
		tg.FileHash{Offset: 2228224, Limit: 131072},
		tg.FileHash{Offset: 2359296, Limit: 131072},
		tg.FileHash{Offset: 2490368, Limit: 131072},
		tg.FileHash{Offset: 2621440, Limit: 131072},
		tg.FileHash{Offset: 2752512, Limit: 131072},
		tg.FileHash{Offset: 2883584, Limit: 131072},
		tg.FileHash{Offset: 3014656, Limit: 131072},
	},
}

type mockHashes struct {
	ranges [][]tg.FileHash
}

func (m mockHashes) Chunk(ctx context.Context, offset, limit int) (chunk, error) {
	panic("implement me")
}

func (m mockHashes) Hashes(ctx context.Context, offset int) ([]tg.FileHash, error) {
	for _, r := range m.ranges {
		last := r[len(r)-1]
		if last.Offset+last.Limit <= offset {
			continue
		}
		return r, nil
	}

	return m.ranges[len(m.ranges)-1], nil
}

func TestVerifier(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		ranges [][]tg.FileHash
		// Hashes returned from CDN redirect, for example.
		predefined []tg.FileHash
		expected   [][]tg.FileHash
	}{
		{"NoPredefined", hashRanges, nil, hashRanges},
		{"Predefined", hashRanges[1:], hashRanges[0], hashRanges},
		{"OnlyPredefined", hashRanges[:1], hashRanges[0], hashRanges[:1]},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := require.New(t)
			m := mockHashes{ranges: test.ranges}
			v := newVerifier(m, test.predefined...)

			hashes := make([]tg.FileHash, 0, len(test.predefined))
			for {
				hash, ok, err := v.next(ctx)
				a.NoError(err)
				if !ok {
					break
				}

				hashes = append(hashes, hash)
			}

			i := 0
			for _, hashRange := range test.expected {
				for _, expected := range hashRange {
					a.Equal(expected, hashes[i])
					i++
				}
			}
		})
	}
}
