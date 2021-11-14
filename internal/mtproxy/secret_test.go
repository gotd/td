package mtproxy

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSecret(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		want    SecretType
		wantErr bool
	}{
		{"Simple", "52a493bdfb90eea55739eabff2d92a14", Simple, false},
		{"Secured", "ddf05fb7acb549be047a7c585116581418", Secured, false},
		{"Secured", "eef05fb7acb549be047a7c585116581418", Secured, false},
		{"TLS-google.com", "ee852380f362a09343efb4690c4e17862e676f6f676c652e636f6d", TLS, false},
		{"TLS-bing.com", "eedf71035a8ed48a623d8e83e66aec4d0562696e672e636f6d", TLS, false},
		{"TLS-yandex.ru", "ee7cea5c13d65f12fd808de70ddcc8d3a979616e6465782e7275", TLS, false},
		{"Bad", "52a493bdfb90eea55739eabff2d92a", 0, true},
		{"Bad", "52a493bdfb90eea55739eabff2d92a1422", 0, true},
		{"Bad", "dd", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := hex.DecodeString(tt.secret)
			require.NoError(t, err)

			got, err := ParseSecret(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				require.Equal(t, tt.want, got.Type)
			}
		})
	}
}
