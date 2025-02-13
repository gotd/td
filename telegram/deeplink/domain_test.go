package deeplink

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		domain  string
		wantErr bool
	}{
		{"a", false},
		{"abcdefghijklmnopqrstuvwxyz123456", false},
		{"Aasdf", false},
		{"asdf0", false},
		{"", true},
		{"asdf_", true},
		{"asd__fg", true},
		{"_asdf", true},
		{"0asdf", true},
		{"9asdf", true},
		{"abcdefghijklmnopqrstuvwxyz1234567", true},
		{"abcdefghijklmnop-qrstuvwxyz", true},
		{"abcdefghijklmnop~qrstuvwxyz", true},
	}
	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			err := ValidateDomain(tt.domain)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
