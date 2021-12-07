package html

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_telegramUnescape(t *testing.T) {
	tests := []struct {
		name string
		b    string
		want string
	}{
		{"NoEscapeCode", "&", "&"},
		{"NoEscapeCode", "&#", "&#"},
		{"UnicodeFlag", "&#127987", string(rune(127987))},
		{"UnicodeFlag", "&#127987;", string(rune(127987))},
		{"UnicodeFlagHex", "&#x1F3f3", string(rune(0x1f3f3))},
		{"UnicodeFlagHex", "&#x1F3f3;", string(rune(0x1f3f3))},
		{"lt", "&lt;", "<"},
		{"lt", "&lt", "<"},
		{"gt", "&gt;", ">"},
		{"gt", "&gt", ">"},
		{"amp", "&amp;", "&"},
		{"amp", "&amp", "&"},
		{"quot", "&quot;", `"`},
		{"quot", "&quot", `"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, []byte(tt.want), telegramUnescape([]byte(tt.b)))
		})
	}
}
