package html

import (
	"strings"
	"testing"

	"github.com/gotd/td/telegram/message/entity"
)

func BenchmarkHTML(b *testing.B) {
	input := `<b>bold</b>, <strong>bold</strong>
<i>italic</i>, <em>italic</em>
	<u>underline</u>, <ins>underline</ins>
	<s>strikethrough</s>, <strike>strikethrough</strike>, <del>strikethrough</del>
	<b>bold <i>italic bold <s>italic bold strikethrough</s> <u>underline italic bold</u></i> bold</b>
	<a href="http://www.example.com/">inline URL</a>
	<a href="tg://user?id=123456789">inline mention of a user</a>
	<code>inline fixed-width code</code>
	<pre>pre-formatted fixed-width code block</pre>
	<pre><code class="language-python">pre-formatted fixed-width code block written in the Python programming language</code></pre>`
	reader := strings.NewReader(input)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reader.Reset(input)
		builder := entity.Builder{}

		if err := HTML(reader, &builder, Options{}); err != nil {
			b.Fatal(err)
		}
	}
}
