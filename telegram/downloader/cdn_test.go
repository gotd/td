package downloader

import (
	"github.com/gotd/td/crypto"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func Test_cdn_decrypt(t *testing.T) {
	testdata := make([]byte, 32)
	tests := []struct {
		name    string
		key, iv []byte
		err     bool
	}{
		{"Bad key", []byte{10}, nil, true},
		{"Bad IV", make([]byte, 32), nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &cdn{
				redirect: &tg.UploadFileCDNRedirect{
					EncryptionKey: test.key,
					EncryptionIv:  test.iv,
				},
			}
			_, err := c.decrypt(testdata, 0, c.redirect)
			if test.err {
				require.Error(t, err)
			}
		})
	}
}

func Test_cdn_hashLookupUnalignedOffset(t *testing.T) {
	c := &cdn{}
	c.cacheHashes([]tg.FileHash{
		{Offset: 128, Limit: 128},
		{Offset: 0, Limit: 128},
		{Offset: 256, Limit: 64},
	})

	type check struct {
		offset int64
		ok     bool
		start  int64
	}
	checks := []check{
		{offset: 0, ok: true, start: 0},
		{offset: 64, ok: true, start: 0},
		{offset: 190, ok: true, start: 128},
		{offset: 319, ok: true, start: 256},
		{offset: 320, ok: false},
	}
	for _, tc := range checks {
		got, ok := c.hash(tc.offset)
		require.Equal(t, tc.ok, ok)
		if !tc.ok {
			continue
		}
		require.Equal(t, tc.start, got.Offset)
	}
}

func Test_cdn_cacheHashesMaintainsSortedOffsets(t *testing.T) {
	c := &cdn{}

	c.cacheHashes([]tg.FileHash{
		{Offset: 256, Limit: 128},
		{Offset: 0, Limit: 128},
		{Offset: 128, Limit: 128},
		{Offset: 128, Limit: 64}, // overwrite existing window, no duplicate offset.
	})

	require.Equal(t, []int64{0, 128, 256}, c.hashOffsets)
	h, ok := c.hash(160)
	require.True(t, ok)
	require.EqualValues(t, 128, h.Offset)
	require.EqualValues(t, 64, h.Limit)
}

func Test_cdn_stateTransitions(t *testing.T) {
	makeHash := func(offset int64) tg.FileHash {
		payload := []byte{byte(offset), byte(offset + 1), byte(offset + 2), byte(offset + 3)}
		return tg.FileHash{
			Offset: offset,
			Limit:  len(payload),
			Hash:   crypto.SHA256(payload),
		}
	}
	makeRedirect := func(offset int64) *tg.UploadFileCDNRedirect {
		hash := makeHash(offset)
		return &tg.UploadFileCDNRedirect{
			DCID:       203,
			FileToken:  []byte{1, 2, 3},
			FileHashes: []tg.FileHash{hash},
		}
	}

	redirectA := makeRedirect(0)
	redirectB := makeRedirect(64)

	tests := []struct {
		name   string
		setup  func(c *cdn)
		action func(c *cdn)
		check  func(t *testing.T, c *cdn, prevRev uint64)
	}{
		{
			name: "MasterToCDNSeedsHashesAndClearsWindows",
			setup: func(c *cdn) {
				c.cacheHashes([]tg.FileHash{makeHash(32)})
				c.cacheWindow(makeHash(32), []byte{1, 2, 3, 4})
			},
			action: func(c *cdn) {
				c.setRedirect(redirectA)
			},
			check: func(t *testing.T, c *cdn, prevRev uint64) {
				require.Equal(t, modeCDN, c.mode)
				require.Same(t, redirectA, c.redirect)
				require.Equal(t, prevRev+1, c.rev)
				_, oldOK := c.hash(32)
				require.False(t, oldOK, "old hash scope should be dropped on redirect")
				seeded, seededOK := c.hash(0)
				require.True(t, seededOK)
				require.EqualValues(t, 0, seeded.Offset)
				require.Nil(t, c.windows)
				require.Nil(t, c.windowsFIFO)
			},
		},
		{
			name: "CDNToMasterClearsRedirectAndCaches",
			setup: func(c *cdn) {
				c.setRedirect(redirectA)
				c.cacheWindow(redirectA.FileHashes[0], []byte{5, 6, 7, 8})
			},
			action: func(c *cdn) {
				c.setMaster()
			},
			check: func(t *testing.T, c *cdn, prevRev uint64) {
				require.Equal(t, modeMaster, c.mode)
				require.Nil(t, c.redirect)
				require.Equal(t, prevRev+1, c.rev)
				_, ok := c.hash(0)
				require.False(t, ok, "master mode should not keep CDN hash cache")
				require.Nil(t, c.windows)
				require.Nil(t, c.windowsFIFO)
			},
		},
		{
			name: "CDNToCDNReplacesScope",
			setup: func(c *cdn) {
				c.setRedirect(redirectA)
				c.cacheWindow(redirectA.FileHashes[0], []byte{9, 9, 9, 9})
			},
			action: func(c *cdn) {
				c.setRedirect(redirectB)
			},
			check: func(t *testing.T, c *cdn, prevRev uint64) {
				require.Equal(t, modeCDN, c.mode)
				require.Same(t, redirectB, c.redirect)
				require.Equal(t, prevRev+1, c.rev)
				_, oldOK := c.hash(0)
				require.False(t, oldOK, "previous redirect hash scope must be invalidated")
				seeded, newOK := c.hash(64)
				require.True(t, newOK)
				require.EqualValues(t, 64, seeded.Offset)
				require.Nil(t, c.windows)
				require.Nil(t, c.windowsFIFO)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &cdn{
				mode: modeMaster,
			}
			if tc.setup != nil {
				tc.setup(c)
			}
			prevRev := c.rev
			tc.action(c)
			tc.check(t, c, prevRev)
		})
	}
}
