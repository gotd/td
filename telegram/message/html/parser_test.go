package html

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
)

type htmlTestCase struct {
	html       string
	msg        string
	entities   func(msg string) []tg.MessageEntityClass
	wantErr    bool
	skipReason string
}

func getEntities(formats ...entity.Formatter) func(msg string) []tg.MessageEntityClass {
	return func(msg string) []tg.MessageEntityClass {
		length := entity.ComputeLength(msg)
		r := make([]tg.MessageEntityClass, len(formats))
		for i := range formats {
			r[i] = formats[i](0, length)
		}
		return r
	}
}

func TestHTML(t *testing.T) {
	runTests := func(tests []htmlTestCase, numericName bool) func(t *testing.T) {
		return func(t *testing.T) {
			for i, test := range tests {
				testName := test.msg
				if numericName || testName == "" {
					testName = fmt.Sprintf("Test%d", i+1)
				}
				t.Run(strings.Title(testName), func(t *testing.T) {
					t.Cleanup(func() {
						if t.Failed() {
							t.Logf("Input: %q", test.html)
						}
					})
					if test.skipReason != "" {
						t.Skip(test.skipReason)
					}
					a := require.New(t)
					b := entity.Builder{}

					err := HTML(strings.NewReader(test.html), &b, Options{})
					if test.wantErr {
						a.Error(err)
						return
					}
					a.NoError(err)

					var (
						msg      string
						entities []tg.MessageEntityClass
					)
					if strings.TrimSpace(test.msg) != test.msg {
						// Complete cuts spaces and fixes entities, but TDLib test expects
						// that it happens after parsing.
						msg, entities = b.Raw()
						entity.SortEntities(entities)
					} else {
						msg, entities = b.Complete()
					}

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

	{
		tests := []htmlTestCase{
			{html: "<b>bold</b>", msg: "bold", entities: getEntities(entity.Bold())},
			{html: "<strong>bold</strong>", msg: "bold", entities: getEntities(entity.Bold())},
			{html: "<i>italic</i>", msg: "italic", entities: getEntities(entity.Italic())},
			{html: "<em>italic</em>", msg: "italic", entities: getEntities(entity.Italic())},
			{html: "<u>underline</u>", msg: "underline", entities: getEntities(entity.Underline())},
			{html: "<ins>underline</ins>", msg: "underline", entities: getEntities(entity.Underline())},
			{html: "<s>strikethrough</s>", msg: "strikethrough", entities: getEntities(entity.Strike())},
			{html: "<strike>strikethrough</strike>", msg: "strikethrough", entities: getEntities(entity.Strike())},
			{html: "<del>strikethrough</del>", msg: "strikethrough", entities: getEntities(entity.Strike())},
			{html: "<code>code</code>", msg: "code", entities: getEntities(entity.Code())},
			{html: "<pre>abc</pre>", msg: "abc", entities: getEntities(entity.Pre(""))},
			{html: `<a href="http://www.example.com/">inline URL</a>`, msg: "inline URL",
				entities: getEntities(entity.TextURL("http://www.example.com/"))},
			{html: `<a href="tg://user?id=123456789">inline mention of a user</a>`, msg: "inline mention of a user",
				entities: getEntities(entity.MentionName(&tg.InputUser{
					UserID: 123456789,
				}))},
			{html: `<pre><code class="language-python">python code</code></pre>`, msg: "python code",
				entities: getEntities(entity.Pre("python"))},
			{html: "<b>&lt;</b>", msg: "<", entities: getEntities(entity.Bold())},
			{html: `<span class="tg-spoiler">spoiler</span>`, msg: "spoiler", entities: getEntities(entity.Spoiler())},
			{html: "<tg-emoji emoji-id=\"5368324170671202286\">👍</tg-emoji>", msg: "👍", entities: getEntities(entity.CustomEmoji(5368324170671202286))},
			{html: "<blockquote expandable>quote</blockquote>", msg: "quote", entities: getEntities(entity.Blockquote(true))},
			{html: "<blockquote>quote</blockquote>", msg: "quote", entities: getEntities(entity.Blockquote(false))},
		}
		t.Run("Common", runTests(tests, false))
	}

	{
		negativeTests := []htmlTestCase{
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

		t.Run("TDLib", runTests(tdlibHTMLTests(), true))
	}
}

func TestIssue525(t *testing.T) {
	test := func(text string, expected []tg.MessageEntityClass) func(t *testing.T) {
		return func(t *testing.T) {
			a := require.New(t)

			b := entity.Builder{}
			p := htmlParser{
				tokenizer: html.NewTokenizer(strings.NewReader(text)),
				builder:   &b,
				attr:      map[string]string{},
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
