package entity

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"github.com/gotd/td/tg"
)

func TestHTML(t *testing.T) {
	type testCase struct {
		html       string
		msg        string
		entities   func(msg string) []tg.MessageEntityClass
		wantErr    bool
		skipReason string
	}

	runTests := func(tests []testCase, numericName bool) func(t *testing.T) {
		return func(t *testing.T) {
			for i, test := range tests {
				msg := test.msg
				if numericName || msg == "" {
					msg = fmt.Sprintf("Test%d", i+1)
				}
				t.Run(strings.Title(msg), func(t *testing.T) {
					t.Logf("Input: %q", test.html)
					if test.skipReason != "" {
						t.Skip(test.skipReason)
					}
					a := require.New(t)
					b := Builder{}

					if err := HTML(strings.NewReader(test.html), &b, nil); test.wantErr {
						a.Error(err)
						return
					} else {
						a.NoError(err)
					}
					if strings.TrimSpace(test.msg) != test.msg {
						t.Skip("Space trimmed by Builder and it's okay")
					}

					msg, entities := b.Complete()
					a.Equal(test.msg, msg)
					if test.entities != nil {
						expect := test.entities(test.msg)
						a.Len(entities, len(expect))
						a.ElementsMatch(expect, entities)
					} else {
						a.Empty(entities)
					}
				})
			}
		}
	}

	getEnities := func(formats ...Formatter) func(msg string) []tg.MessageEntityClass {
		return func(msg string) []tg.MessageEntityClass {
			length := ComputeLength(msg)
			r := make([]tg.MessageEntityClass, len(formats))
			for i := range formats {
				r[i] = formats[i](0, length)
			}
			return r
		}
	}

	{
		tests := []testCase{
			{html: "<b>bold</b>", msg: "bold", entities: getEnities(Bold())},
			{html: "<strong>bold</strong>", msg: "bold", entities: getEnities(Bold())},
			{html: "<i>italic</i>", msg: "italic", entities: getEnities(Italic())},
			{html: "<em>italic</em>", msg: "italic", entities: getEnities(Italic())},
			{html: "<u>underline</u>", msg: "underline", entities: getEnities(Underline())},
			{html: "<ins>underline</ins>", msg: "underline", entities: getEnities(Underline())},
			{html: "<s>strikethrough</s>", msg: "strikethrough", entities: getEnities(Strike())},
			{html: "<strike>strikethrough</strike>", msg: "strikethrough", entities: getEnities(Strike())},
			{html: "<del>strikethrough</del>", msg: "strikethrough", entities: getEnities(Strike())},
			{html: "<code>code</code>", msg: "code", entities: getEnities(Code())},
			{html: "<pre>abc</pre>", msg: "abc", entities: getEnities(Pre(""))},
			{html: `<a href="http://www.example.com/">inline URL</a>`, msg: "inline URL",
				entities: getEnities(TextURL("http://www.example.com/"))},
			{html: `<a href="tg://user?id=123456789">inline mention of a user</a>`, msg: "inline mention of a user",
				entities: getEnities(MentionName(&tg.InputUser{
					UserID: 123456789,
				}))},
			{html: `<pre><code class="language-python">python code</code></pre>`, msg: "python code",
				entities: getEnities(Code(), Pre("python"))},
			{html: "<b>&lt;</b>", msg: "<", entities: getEnities(Bold())},
		}
		t.Run("Common", runTests(tests, false))
	}

	{
		negativeTests := []testCase{
			{html: "&#57311;", wantErr: true},
			{html: "&#xDFDF;", wantErr: true},
			{html: "&#xDFDF", wantErr: true},
			{html: "🏟 🏟&lt;<abacaba", wantErr: true},
			{html: "🏟 🏟&lt;<abac aba>", wantErr: true},
			{html: "🏟 🏟&lt;<abac>", wantErr: true},
			{html: "🏟 🏟&lt;<i   =aba>", wantErr: true},
			{html: "🏟 🏟&lt;<i    aba>", wantErr: true},
			{html: "🏟 🏟&lt;<i    aba  =  ", wantErr: true},
			{html: "🏟 🏟&lt;<i    aba  =  190azAz-.,", wantErr: true},
			{html: "🏟 🏟&lt;<i    aba  =  \"&lt;&gt;&quot;>", wantErr: true},
			{html: "🏟 🏟&lt;<i    aba  =  \\'&lt;&gt;&quot;>", wantErr: true},
			{html: "🏟 🏟&lt;</", wantErr: true},
			{html: "🏟 🏟&lt;<b></b></", wantErr: true},
			{html: "🏟 🏟&lt;<i>a</i   ", wantErr: true},
			{html: "🏟 🏟&lt;<i>a</em   >", wantErr: true},
		}
		// FIXME(tdakkota): sanitize HTML
		_ = negativeTests

		entities := func(e ...tg.MessageEntityClass) func(msg string) []tg.MessageEntityClass {
			return func(msg string) []tg.MessageEntityClass {
				return e
			}
		}
		tdlibCompat := []testCase{
			{"", "", nil, false, ""},
			{"➡️ ➡️", "➡️ ➡️", nil, false, ""},
			{
				"&lt;&gt;&amp;&quot;&laquo;&raquo;&#12345678;",
				"<>&\"&laquo;&raquo;&#12345678;",
				nil,
				false,
				"Custom escape is incomplete",
			},

			{
				"➡️ ➡️<i>➡️ ➡️</i>",
				"➡️ ➡️➡️ ➡️",
				entities(&tg.MessageEntityItalic{Offset: 5, Length: 5}),
				false,
				"",
			},
			{
				"➡️ ➡️<em>➡️ ➡️</em>", "➡️ ➡️➡️ ➡️",
				entities(&tg.MessageEntityItalic{Offset: 5, Length: 5}),
				false,
				"",
			},
			{
				"➡️ ➡️<b>➡️ ➡️</b>",
				"➡️ ➡️➡️ ➡️",
				entities(&tg.MessageEntityBold{Offset: 5, Length: 5}),
				false,
				"",
			},
			{
				"➡️ ➡️<stro" +
					"ng>➡️ ➡️</strong>",
				"➡️ ➡️➡️ ➡️",
				entities(&tg.MessageEntityBold{Offset: 5, Length: 5}),
				false,
				"",
			},
			{
				"➡️ ➡️<u>➡️ ➡️</u>",
				"➡️ ➡️➡️ ➡️",
				entities(&tg.MessageEntityUnderline{Offset: 5, Length: 5}),
				false,
				"",
			},
			{
				"➡️ ➡️<ins>➡️ ➡️</ins>",
				"➡️ ➡️➡️ ➡️",
				entities(&tg.MessageEntityUnderline{Offset: 5, Length: 5}),
				false,
				"",
			},
			{
				"➡️ ➡️<s>➡️ ➡️</s>",
				"➡️ ➡️➡️ ➡️",
				entities(&tg.MessageEntityStrike{Offset: 5, Length: 5}),
				false,
				"",
			},
			{
				"➡️ ➡️<strike>➡️ ➡️</strike>",
				"➡️ ➡️➡️ ➡️",
				entities(&tg.MessageEntityStrike{Offset: 5, Length: 5}),
				false,
				"",
			},
			{
				"➡️ ➡️<del>➡️ ➡️</del>",
				"➡️ ➡️➡️ ➡️",
				entities(&tg.MessageEntityStrike{Offset: 5, Length: 5}),
				false,
				"",
			},
			{
				"➡️ ➡️<i>➡️ ➡️</i><b>➡️ ➡️</b>",
				"➡️ ➡️➡️ ➡️➡️ ➡️",
				entities(
					&tg.MessageEntityItalic{Offset: 5, Length: 5},
					&tg.MessageEntityBold{Offset: 10, Length: 5},
				),
				false,
				"",
			},

			{
				"🏟 🏟<i>🏟 &lt🏟</i>",
				"🏟 🏟🏟 <🏟",
				entities(&tg.MessageEntityItalic{Offset: 5, Length: 6}),
				false,
				"",
			},
			{
				"🏟 🏟<i>🏟 &gt;<b aba   =   caba>&lt🏟</b></i>",
				"🏟 🏟🏟 ><🏟",
				entities(
					&tg.MessageEntityItalic{Offset: 5, Length: 7},
					&tg.MessageEntityBold{Offset: 9, Length: 3},
				),
				false,
				"",
			},
			{
				"🏟 🏟&lt;<i    aba  =  190azAz-.   >a</i>",
				"🏟 🏟<a",
				entities(&tg.MessageEntityItalic{Offset: 6, Length: 1}),
				false,
				"",
			},
			{
				"🏟 🏟&lt;<i    aba  =  190azAz-.>a</i>",
				"🏟 🏟<a",
				entities(&tg.MessageEntityItalic{Offset: 6, Length: 1}),
				false,
				"",
			},
			{
				"🏟 🏟&lt;<i    aba  =  \"&lt;&gt;&quot;\">a</i>",
				"🏟 🏟<a",
				entities(&tg.MessageEntityItalic{Offset: 6, Length: 1}),
				false,
				"",
			},
			{
				"🏟 🏟&lt;<i    aba  =  '&lt;&gt;&quot;'>a</i>",
				"🏟 🏟<a",
				entities(&tg.MessageEntityItalic{Offset: 6, Length: 1}),
				false,
				"",
			},
			{
				"🏟 🏟&lt;<i    aba  =  '&lt;&gt;&quot;'>a</>",
				"🏟 🏟<a",
				entities(&tg.MessageEntityItalic{Offset: 6, Length: 1}),
				false,
				"",
			},
			{
				"🏟 🏟&lt;<i>🏟 🏟&lt;</>",
				"🏟 🏟<🏟 🏟<",
				entities(&tg.MessageEntityItalic{Offset: 6, Length: 6}),
				false,
				"",
			},

			{
				"🏟 🏟&lt;<i>a</    >",
				"🏟 🏟<a",
				entities(&tg.MessageEntityItalic{Offset: 6, Length: 1}),
				false,
				"",
			},
			{
				"🏟 🏟&lt;<i>a</i   >",
				"🏟 🏟<a",
				entities(&tg.MessageEntityItalic{Offset: 6, Length: 1}),
				false,
				"",
			},
			// Empty entity.
			{
				"🏟 🏟&lt;<b></b>",
				"🏟 🏟<",
				nil,
				false,
				"",
			},
			// Space handling.
			{
				"<i>\t</i>",
				"\t",
				entities(&tg.MessageEntityItalic{Offset: 0, Length: 1}),
				false,
				"",
			},
			{
				"<i>\r</i>",
				"\r",
				entities(&tg.MessageEntityItalic{Offset: 0, Length: 1}),
				false,
				"",
			},
			{
				"<i>\n</i>",
				"\n",
				entities(&tg.MessageEntityItalic{Offset: 0, Length: 1}),
				false,
				"",
			},
			{
				"<a href=telegram.org>\t</a>",
				"\t",
				entities(&tg.MessageEntityTextURL{Offset: 0, Length: 1, URL: "http://telegram.org/"}),
				false,
				"",
			},
			{
				"<a href=telegram.org>\r</a>",
				"\r",
				entities(&tg.MessageEntityTextURL{Offset: 0, Length: 1, URL: "http://telegram.org/"}),
				false,
				"",
			},
			{
				"<a href=telegram.org>\n</a>",
				"\n",
				entities(&tg.MessageEntityTextURL{Offset: 0, Length: 1, URL: "http://telegram.org/"}),
				false,
				"",
			},
			{
				"<code><i><b> </b></i></code><i><b><code> </code></b></i>",
				"  ",
				entities(
					&tg.MessageEntityCode{Offset: 0, Length: 1},
					&tg.MessageEntityBold{Offset: 0, Length: 1},
					&tg.MessageEntityItalic{Offset: 0, Length: 1},
					&tg.MessageEntityCode{Offset: 1, Length: 1},
					&tg.MessageEntityBold{Offset: 1, Length: 1},
					&tg.MessageEntityItalic{Offset: 1, Length: 1}),
				false,
				"",
			},
			{
				"<i><b> </b> <code> </code></i>",
				"   ",
				entities(
					&tg.MessageEntityItalic{Offset: 0, Length: 3},
					&tg.MessageEntityBold{Offset: 0, Length: 1},
					&tg.MessageEntityCode{Offset: 2, Length: 1},
				),
				false,
				"",
			},
			{
				"<a href=telegram.org> </a>",
				" ",
				entities(&tg.MessageEntityTextURL{Offset: 0, Length: 1, URL: "http://telegram.org/"}),
				false,
				"",
			},
			{
				"<a href  =\"telegram.org\"   > </a>",
				" ",
				entities(&tg.MessageEntityTextURL{Offset: 0, Length: 1, URL: "http://telegram.org/"}),
				false,
				"",
			},
			{
				"<a   href=  'telegram.org'   > </a>",
				" ",
				entities(&tg.MessageEntityTextURL{Offset: 0, Length: 1, URL: "http://telegram.org/"}),
				false,
				"",
			},
			{
				"<a   href=  'telegram.org?&lt;'   > </a>",
				" ",
				entities(&tg.MessageEntityTextURL{Offset: 0, Length: 1, URL: "http://telegram.org/?<"}),
				false,
				"",
			},
			// URL handling
			{
				"<a>telegram.org </a>",
				"telegram.org ",
				nil,
				false,
				"URL parsing from text is incomplete",
			},
			{
				"<a>telegram.org</a>", "telegram.org",
				entities(&tg.MessageEntityTextURL{
					Offset: 0,
					Length: 12,
					URL:    "http://telegram.org/",
				}),
				false,
				"URL parsing from text is incomplete",
			},
			{
				"<a>https://telegram.org/asdsa?asdasdwe#12e3we</a>",
				"https://telegram.org/asdsa?asdasdwe#12e3we",
				entities(&tg.MessageEntityTextURL{
					Offset: 0,
					Length: 42,
					URL:    "https://telegram.org/asdsa?asdasdwe#12e3we",
				}),
				false,
				"URL parsing from text is incomplete",
			},
			// <pre> and <code> handling
			{
				"🏟 🏟&lt;<pre  >🏟 🏟&lt;</>",
				"🏟 🏟<🏟 🏟<",
				entities(&tg.MessageEntityPre{Offset: 6, Length: 6}),
				false,
				"",
			},
			{
				"🏟 🏟&lt;<code >🏟 🏟&lt;</>",
				"🏟 🏟<🏟 🏟<",
				entities(&tg.MessageEntityCode{Offset: 6, Length: 6}),
				false,
				"",
			},
			{
				"🏟 🏟&lt;<pre><code>🏟 🏟&lt;</code></>",
				"🏟 🏟<🏟 🏟<",
				entities(
					&tg.MessageEntityPre{Offset: 6, Length: 6},
					&tg.MessageEntityCode{Offset: 6, Length: 6},
				),
				false,
				"",
			},
			{
				"🏟 🏟&lt;<pre><code class=\"language-\">🏟 🏟&lt;</code></>",
				"🏟 🏟<🏟 🏟<",
				entities(
					&tg.MessageEntityPre{Offset: 6, Length: 6},
					&tg.MessageEntityCode{Offset: 6, Length: 6},
				),
				false,
				"",
			},
			{
				"🏟 🏟&lt;<pre><code class=\"language-fift\">🏟 🏟&lt;</></>",
				"🏟 🏟<🏟 🏟<",
				entities(&tg.MessageEntityPre{Offset: 6, Length: 6, Language: "fift"}),
				false,
				"<pre> and <code> shrink is incomplete",
			},
			{
				"🏟 🏟&lt;<code class=\"language-fift\"><pre>🏟 🏟&lt;</></>",
				"🏟 🏟<🏟 🏟<",
				entities(&tg.MessageEntityPre{Offset: 6, Length: 6, Language: "fift"}),
				false,
				"<pre> and <code> shrink is incomplete",
			},
			{
				"🏟 🏟&lt;<pre><code class=\"language-fift\">🏟 🏟&lt;</> </>",
				"🏟 🏟<🏟 🏟< ",
				entities(
					&tg.MessageEntityPre{Offset: 6, Length: 7},
					&tg.MessageEntityCode{Offset: 6, Length: 6},
				),
				false,
				"<pre> and <code> shrink is incomplete",
			},
			{
				"🏟 🏟&lt;<pre> <code class=\"language-fift\">🏟 🏟&lt;</></>",
				"🏟 🏟< 🏟 🏟<",
				entities(
					&tg.MessageEntityPre{Offset: 6, Length: 7},
					&tg.MessageEntityCode{Offset: 7, Length: 6},
				),
				false,
				"BUG: TDLib does not add language tag for some reason",
			},
		}
		t.Run("TDLib", runTests(tdlibCompat, true))
	}
}

func TestIssue525(t *testing.T) {
	test := func(text string, expected []tg.MessageEntityClass) func(t *testing.T) {
		return func(t *testing.T) {
			a := require.New(t)

			b := Builder{}
			p := htmlParser{
				tokenizer:    html.NewTokenizer(strings.NewReader(text)),
				builder:      &b,
				attr:         map[string]string{},
				userResolver: nil,
			}

			a.NoError(p.parse())
			_, entities := b.Complete()
			a.Equal(expected, entities)
		}
	}

	t.Run("Ru", test(`Строка
<i>Строка текста курсивом</i>

Обычный текст с <a href="https://google.com">Ссылкой</a> внутри, и
ещё одна ссылка - <a href="https://go.dev">Здесь</a>.

Ещё одна строка.
`,
		[]tg.MessageEntityClass{
			&tg.MessageEntityItalic{
				Offset: 7,
				Length: 22,
			},
			&tg.MessageEntityTextURL{
				Offset: 47,
				Length: 7,
				URL:    "https://google.com",
			},
			&tg.MessageEntityTextURL{
				Offset: 83,
				Length: 5,
				URL:    "https://go.dev",
			},
		}),
	)
	t.Run("En", test(`Line
<i>Italic line of text</i>

Normal line of text with <a href="https://google.com">Link</a> inside, and
another link now - <a href="https://go.dev">Here</a>.

One more line.
`,
		[]tg.MessageEntityClass{
			&tg.MessageEntityItalic{
				Offset: 5,
				Length: 19,
			},
			&tg.MessageEntityTextURL{
				Offset: 51,
				Length: 4,
				URL:    "https://google.com",
			},
			&tg.MessageEntityTextURL{
				Offset: 87,
				Length: 4,
				URL:    "https://go.dev",
			},
		}),
	)
}
