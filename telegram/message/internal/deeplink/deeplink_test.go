package deeplink

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type testCase struct {
	link    DeepLink
	input   string
	wantErr bool
}

func join(arg string) DeepLink {
	return DeepLink{
		Type: Join,
		Args: map[string][]string{
			"invite": {arg},
		},
	}
}
func resolve(arg string) DeepLink {
	return DeepLink{
		Type: Resolve,
		Args: map[string][]string{
			"domain": {arg},
		},
	}
}

func joinSuite() map[string][]testCase {
	expect := join("AAAAAAAAAAAAAAAAAA")
	return map[string][]testCase{
		"Test": {
			{expect, `t.me/joinchat/AAAAAAAAAAAAAAAAAA`, false},
			{expect, `t.me/joinchat/AAAAAAAAAAAAAAAAAA/`, false},
			{expect, `t.me/+AAAAAAAAAAAAAAAAAA`, false},
			{expect, `t.me/+AAAAAAAAAAAAAAAAAA/`, false},
			{expect, `t.me/  +AAAAAAAAAAAAAAAAAA/`, false},
			{expect, `https://t.me/joinchat/AAAAAAAAAAAAAAAAAA`, false},
			{expect, `https://t.me/joinchat/AAAAAAAAAAAAAAAAAA/`, false},
			{expect, `tg:join?invite=AAAAAAAAAAAAAAAAAA`, false},
			{expect, `tg://join?invite=AAAAAAAAAAAAAAAAAA`, false},

			{DeepLink{}, `https://t.co/joinchat/AAAAAAAAAAAAAAAAAA`, true},
			{DeepLink{}, `rt://join?invite=AAAAAAAAAAAAAAAAAA`, true},
		},
		"TDLib": {
			// t.me/+<hash>
			// Positive
			{join("aba%20aba"), "t.me/+aba%20aba", false},
			{join("aba0aba"), "t.me/+aba%30aba", false},
			{join("123456a"), "t.me/+123456a", false},
			{join("12345678901"), "t.me/%2012345678901", false},
			// Negative
			{DeepLink{}, "t.me/+?invite=abcdef", true},
			{DeepLink{}, "t.me/+", true},
			{DeepLink{}, "t.me/+/abcdef", true},
			{DeepLink{}, "t.me/ ?/abcdef", true},
			{DeepLink{}, "t.me/+?abcdef", true},
			{DeepLink{}, "t.me/+#abcdef", true},
			{DeepLink{}, "t.me/ /123456/123123/12/31/a/s//21w/?asdas#test", true},

			// t.me/joinchat/<hash>
			// Positive
			{join("abacaba"), "t.me/joinchat/abacaba", false},
			{join("aba%20aba"), "t.me/joinchat/aba%20aba", false},
			{join("aba0aba"), "t.me/joinchat/aba%30aba", false},
			{join("123456a"), "t.me/joinchat/123456a", false},
			{join("12345678901"), "t.me/joinchat/12345678901", false},
			{join("123456"), "t.me/joinchat/123456", false},
			{join("123456"), "t.me/joinchat/123456/123123/12/31/a/s//21w/?asdas#test", false},
			// Negative
			{DeepLink{}, "t.me/joinchat?invite=abcdef", true},
			{DeepLink{}, "t.me/joinchat", true},
			{DeepLink{}, "t.me/joinchat/", true},
			{DeepLink{}, "t.me/joinchat//abcdef", true},
			{DeepLink{}, "t.me/joinchat?/abcdef", true},
			{DeepLink{}, "t.me/joinchat/?abcdef", true},
			{DeepLink{}, "t.me/joinchat/#abcdef", true},
		},
	}
}

func resolveSuite() map[string][]testCase {
	expect := resolve("gotd_ru")
	return map[string][]testCase{
		"Test": {
			{expect, `t.me/gotd_ru`, false},
			{expect, `t.me/gotd_ru/`, false},
			{expect, `https://t.me/gotd_ru`, false},
			{expect, `https://t.me/gotd_ru/`, false},
			{expect, `tg:resolve?domain=gotd_ru`, false},
			{expect, `tg:resolve?&&&&&&&domain=gotd_ru`, false},
			{expect, `tg://resolve?domain=gotd_ru`, false},

			{DeepLink{}, `https://t.co/gotd_ru`, true},
			{DeepLink{}, `rt://join?invite=AAAAAAAAAAAAAAAAAA`, true},
		},
		"TDLib": {
			// t.me/<domain>
			// Positive
			{resolve("a"), "t.me/a", false},
			{resolve("abcdefghijklmnopqrstuvwxyz123456"), "t.me/abcdefghijklmnopqrstuvwxyz123456", false},
			{resolve("Aasdf"), "t.me/Aasdf", false},
			{resolve("asdf0"), "t.me/asdf0", false},
			{resolve("username"), "t.me/username/0/a//s/as?gam=asd", false},
			{resolve("username"), "t.me/username/aasdas?test=1", false},
			{resolve("username"), "t.me/username/0", false},
			{resolve("telecram"), "https://telegram.dog/tele%63ram", false},
			// Negative
			{DeepLink{}, "t.me/abcdefghijklmnopqrstuvwxyz1234567", true},
			{DeepLink{}, "t.me/abcdefghijklmnop-qrstuvwxyz", true},
			{DeepLink{}, "t.me/abcdefghijklmnop~qrstuvwxyz", true},
			{DeepLink{}, "t.me/_asdf", true},
			{DeepLink{}, "t.me/0asdf", true},
			{DeepLink{}, "t.me/9asdf", true},
			{DeepLink{}, "t.me/asdf_", true},
			{DeepLink{}, "t.me/asd__fg", true},
			{DeepLink{}, "t.me//username", true},
		},
	}
}

var typeSuites = map[string]map[string][]testCase{
	"Join":    joinSuite(),
	"Resolve": resolveSuite(),
}

func TestParseDeeplink(t *testing.T) {
	runSuite := func(suite []testCase) func(t *testing.T) {
		return func(t *testing.T) {
			for i, test := range suite {
				t.Run(fmt.Sprintf("Test%d (%s)", i, test.input), func(t *testing.T) {
					a := require.New(t)
					d, err := Parse(test.input)

					if test.wantErr {
						a.Error(err, test.input)
					} else {
						a.NoError(err, test.input)
						a.Equal(test.link, d, test.input)
					}
				})
			}
		}
	}

	for typeName, typeSuite := range typeSuites {
		t.Run(typeName, func(t *testing.T) {
			for suiteName, suite := range typeSuite {
				t.Run(suiteName, runSuite(suite))
			}
		})
	}
}
