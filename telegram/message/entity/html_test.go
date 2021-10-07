package entity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func TestHTML(t *testing.T) {
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

	tests := []struct {
		html     string
		msg      string
		entities func(msg string) []tg.MessageEntityClass
	}{
		{"<b>bold</b>", "bold", getEnities(Bold())},
		{"<strong>bold</strong>", "bold", getEnities(Bold())},
		{"<i>italic</i>", "italic", getEnities(Italic())},
		{"<em>italic</em>", "italic", getEnities(Italic())},
		{"<u>underline</u>", "underline", getEnities(Underline())},
		{"<ins>underline</ins>", "underline", getEnities(Underline())},
		{"<s>strikethrough</s>", "strikethrough", getEnities(Strike())},
		{"<strike>strikethrough</strike>", "strikethrough", getEnities(Strike())},
		{"<del>strikethrough</del>", "strikethrough", getEnities(Strike())},
		{"<code>code</code>", "code", getEnities(Code())},
		{"<pre>abc</pre>", "abc", getEnities(Code())},
		{`<a href="http://www.example.com/">inline URL</a>`, "inline URL",
			getEnities(TextURL("http://www.example.com/"))},
		{`<a href="tg://user?id=123456789">inline mention of a user</a>`, "inline mention of a user",
			getEnities(MentionName(&tg.InputUser{
				UserID: 123456789,
			}))},
		{`<pre><code class="language-python">python code</code></pre>`, "python code",
			getEnities(Pre("python"), Code())},
		{"<b>&lt;</b>", "<", getEnities(Bold())},
	}

	for _, test := range tests {
		t.Run(strings.Title(test.msg), func(t *testing.T) {
			a := require.New(t)
			b := Builder{}
			a.NoError(HTML(strings.NewReader(test.html), &b, nil))

			msg, entities := b.Complete()
			a.Equal(test.msg, msg)
			a.Equal(test.entities(test.msg), entities)
		})
	}
}
