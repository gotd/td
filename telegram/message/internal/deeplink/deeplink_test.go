package deeplink

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDeeplink(t *testing.T) {
	expect := DeepLink{
		Type: Join,
		Args: map[string][]string{
			"invite": {"AAAAAAAAAAAAAAAAAA"},
		},
	}
	tests := []struct {
		link    DeepLink
		input   string
		wantErr bool
	}{
		{expect, `t.me/joinchat/AAAAAAAAAAAAAAAAAA`, false},
		{expect, `t.me/joinchat/AAAAAAAAAAAAAAAAAA/`, false},
		{expect, `t.me/+AAAAAAAAAAAAAAAAAA`, false},
		{expect, `t.me/+AAAAAAAAAAAAAAAAAA/`, false},
		{expect, `https://t.me/joinchat/AAAAAAAAAAAAAAAAAA`, false},
		{expect, `https://t.me/joinchat/AAAAAAAAAAAAAAAAAA/`, false},
		{expect, `tg:join?invite=AAAAAAAAAAAAAAAAAA`, false},
		{expect, `tg://join?invite=AAAAAAAAAAAAAAAAAA`, false},

		{DeepLink{}, `https://t.co/joinchat/AAAAAAAAAAAAAAAAAA`, true},
		{DeepLink{}, `rt://join?invite=AAAAAAAAAAAAAAAAAA`, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			a := require.New(t)
			d, err := Parse(test.input)

			if test.wantErr {
				a.Error(err)
			} else {
				a.NoError(err)
				a.Equal(test.link, d)
			}
		})
	}
}
