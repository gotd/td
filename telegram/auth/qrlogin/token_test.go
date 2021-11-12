package qrlogin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseTokenURL(t *testing.T) {

	tests := []struct {
		name    string
		u       string
		want    Token
		wantErr bool
	}{
		{
			"Valid",
			"tg://login?token=AQL0cY5hVg_D1OqESdYnJVg5845qbd8FiOLpUUeyvcb28g==",
			Token{
				token: []uint8{
					0x1, 0x2, 0xf4, 0x71, 0x8e, 0x61,
					0x56, 0xf, 0xc3, 0xd4, 0xea, 0x84,
					0x49, 0xd6, 0x27, 0x25, 0x58, 0x39,
					0xf3, 0x8e, 0x6a, 0x6d, 0xdf, 0x5,
					0x88, 0xe2, 0xe9, 0x51, 0x47, 0xb2,
					0xbd, 0xc6, 0xf6, 0xf2,
				},
				expires: time.Unix(0, 0),
			},
			false,
		},
		{"InvalidSchema", "vk://login", Token{}, true},
		{"InvalidPath", "tg://aboba", Token{}, true},
		{"NoToken", "tg://login", Token{}, true},
		{"InvalidBase64", "tg://login?token=A", Token{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)

			got, err := ParseTokenURL(tt.u)
			if tt.wantErr {
				a.Error(err)
			} else {
				a.Equal(tt.want, got)
				a.NoError(err)
				a.Equal(tt.want.URL(), tt.u)
			}
		})
	}
}
