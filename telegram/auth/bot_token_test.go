package auth

import (
	"strings"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
)

func TestParseBotToken(t *testing.T) {
	const secret = "AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw"
	for _, tt := range []struct {
		name  string
		token string
		id    int64
		ok    bool
	}{
		{"OK", "123456:" + secret, 123456, true},
		{"MaxID", "18014398509481983:x", 1<<54 - 1, true},
		{"SecretWithColon", "1:a:b", 1, true},
		{"Max length", "1:" + strings.Repeat("a", 78), 1, true},
		{"TooLong", "1:" + strings.Repeat("a", 79), 0, false},
		{"Empty", "", 0, false},
		{"NoColon", "123456", 0, false},
		{"Slash", "123456:" + secret + "/test", 0, false},
		{"LeadingZero", "0123:" + secret, 0, false},
		{"Zero", "0:" + secret, 0, false},
		{"EmptyID", ":" + secret, 0, false},
		{"Negative", "-1:" + secret, 0, false},
		{"Plus", "+1:" + secret, 0, false},
		{"NotNumber", "12a:" + secret, 0, false},
		{"IDTooLarge", "18014398509481984:" + secret, 0, false},
		{"IDOverflow", "99999999999999999999:" + secret, 0, false},
		{"EmptySecret", "123456:", 0, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tok, err := ParseBotToken(tt.token)
			if !tt.ok {
				require.Error(t, err)
				require.True(t, errors.Is(err, ErrInvalidBotToken))
				if s := strings.SplitN(tt.token, ":", 2); len(s) == 2 && s[1] != "" {
					require.NotContains(t, err.Error(), s[1], "error must not leak the secret")
				}
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.id, tok.ID)
			require.Equal(t, tt.token, tok.String())
		})
	}
}
