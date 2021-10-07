package session

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/session/tdesktop"
	"github.com/nnqq/td/telegram/dcs"
)

func TestTDesktopSession(t *testing.T) {
	key := crypto.Key{}
	keyID := key.ID()

	tests := []struct {
		name    string
		account tdesktop.Account
		want    *Data
		wantErr bool
	}{
		{"OK", tdesktop.Account{
			Authorization: tdesktop.MTPAuthorization{
				MainDC: 2,
				Keys: map[int]crypto.Key{
					2: key,
				},
			},
		}, &Data{
			DC:        2,
			Addr:      findDCAddr(dcs.Prod().Options, 2),
			AuthKey:   key[:],
			AuthKeyID: keyID[:],
		}, false},
		{"UnknownDC", tdesktop.Account{
			Authorization: tdesktop.MTPAuthorization{
				MainDC: 200,
				Keys: map[int]crypto.Key{
					200: key,
				},
			},
		}, nil, true},
		{"NoKey", tdesktop.Account{}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)
			got, err := TDesktopSession(tt.account)
			if tt.wantErr {
				a.Nil(got)
				a.Error(err)
			} else {
				a.Equal(tt.want, got)
				a.NoError(err)
			}
		})
	}
}
