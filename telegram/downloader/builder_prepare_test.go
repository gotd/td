package downloader

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuilder_preparePaths(t *testing.T) {
	newMock := func() *mock {
		return &mock{data: []byte("hello")}
	}

	tests := []struct {
		name            string
		build           func() *Builder
		wantCDNSchema   bool
		wantMasterAllow bool
		wantInlineCDN   bool
		wantOuterVerify bool
	}{
		{
			name: "LegacyFastPathWhenFlagUnset",
			build: func() *Builder {
				return NewDownloader().Download(newMock(), nil)
			},
			wantCDNSchema:   false,
			wantMasterAllow: false,
			wantOuterVerify: false,
		},
		{
			name: "LegacyFastPathWhenNoProvider",
			build: func() *Builder {
				m := newMock()
				return NewDownloader().WithAllowCDN(true).Download(&noCDNClient{base: m}, nil)
			},
			wantCDNSchema:   false,
			wantMasterAllow: false,
			wantOuterVerify: false,
		},
		{
			name: "CDNPathDefaultEnablesInlineVerify",
			build: func() *Builder {
				return NewDownloader().WithAllowCDN(true).Download(newMock(), nil)
			},
			wantCDNSchema:   true,
			wantMasterAllow: true,
			wantInlineCDN:   true,
			wantOuterVerify: false,
		},
		{
			name: "CDNPathExplicitVerifyKeepsOuterVerifier",
			build: func() *Builder {
				return NewDownloader().WithAllowCDN(true).Download(newMock(), nil).WithVerify(true)
			},
			wantCDNSchema:   true,
			wantMasterAllow: true,
			wantInlineCDN:   false,
			wantOuterVerify: true,
		},
		{
			name: "CDNPathExplicitDisableKeepsInlineVerify",
			build: func() *Builder {
				return NewDownloader().WithAllowCDN(true).Download(newMock(), nil).WithVerify(false)
			},
			wantCDNSchema:   true,
			wantMasterAllow: true,
			wantInlineCDN:   true,
			wantOuterVerify: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			prepared, closeCDN, err := tc.build().prepare()
			require.NoError(t, err)
			if closeCDN != nil {
				defer func() {
					require.NoError(t, closeCDN())
				}()
			}

			require.Equal(t, tc.wantOuterVerify, prepared.verify)

			if tc.wantCDNSchema {
				s, ok := prepared.schema.(*cdn)
				require.True(t, ok)
				require.Equal(t, tc.wantMasterAllow, s.master.allowCDN)
				require.Equal(t, tc.wantInlineCDN, s.verify)
				return
			}

			s, ok := prepared.schema.(master)
			require.True(t, ok)
			require.Equal(t, tc.wantMasterAllow, s.allowCDN)
		})
	}
}
