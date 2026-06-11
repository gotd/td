package rich

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func parseHTML(t *testing.T, s string) []tg.PageBlockClass {
	t.Helper()
	blocks, err := ParseHTML(strings.NewReader(s))
	require.NoError(t, err)
	return blocks
}

func parseMarkdown(t *testing.T, s string) []tg.PageBlockClass {
	t.Helper()
	blocks, err := ParseMarkdown(strings.NewReader(s))
	require.NoError(t, err)
	return blocks
}

func TestParseHTML(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   string
		want []tg.PageBlockClass
	}{
		{
			name: "Paragraph",
			in:   "<p>Hello <b>world</b></p>",
			want: []tg.PageBlockClass{Paragraph(Concat(Plain("Hello "), Bold(Plain("world"))))},
		},
		{
			name: "Headings",
			in:   "<h1>One</h1><h3>Three</h3>",
			want: []tg.PageBlockClass{Heading1(Plain("One")), Heading3(Plain("Three"))},
		},
		{
			name: "Divider",
			in:   "<p>a</p><hr><p>b</p>",
			want: []tg.PageBlockClass{Paragraph(Plain("a")), Divider(), Paragraph(Plain("b"))},
		},
		{
			name: "Subscript and superscript",
			in:   "<p>H<sub>2</sub>O and E=mc<sup>2</sup></p>",
			want: []tg.PageBlockClass{Paragraph(Concat(
				Plain("H"), Subscript(Plain("2")), Plain("O and E=mc"), Superscript(Plain("2")),
			))},
		},
		{
			name: "Marked and spoiler and math",
			in:   `<p><mark>hi</mark> <tg-spoiler>secret</tg-spoiler> <tg-math>x^2</tg-math></p>`,
			want: []tg.PageBlockClass{Paragraph(Concat(
				Marked(Plain("hi")), Plain(" "), Spoiler(Plain("secret")), Plain(" "), Math("x^2"),
			))},
		},
		{
			name: "Anchor and anchor link",
			in:   `<p><a name="top">Top</a> <a href="#top">back</a></p>`,
			want: []tg.PageBlockClass{Paragraph(Concat(
				Anchor(Plain("Top"), "top"), Plain(" "), AnchorLink(Plain("back"), "top"),
			))},
		},
		{
			name: "Link and email",
			in:   `<p><a href="https://go.dev">Go</a> <a href="mailto:a@b.c">mail</a></p>`,
			want: []tg.PageBlockClass{Paragraph(Concat(
				URL(Plain("Go"), "https://go.dev", 0), Plain(" "), Email(Plain("mail"), "a@b.c"),
			))},
		},
		{
			name: "Preformatted",
			in:   `<pre><code class="language-go">x := 1</code></pre>`,
			want: []tg.PageBlockClass{Preformatted(Plain("x := 1"), "go")},
		},
		{
			name: "Blockquote",
			in:   "<blockquote><p>quoted</p></blockquote>",
			want: []tg.PageBlockClass{Blockquote(Plain("quoted"), Empty())},
		},
		{
			name: "Unordered list",
			in:   "<ul><li>one</li><li>two</li></ul>",
			want: []tg.PageBlockClass{List(ListItem(Plain("one")), ListItem(Plain("two")))},
		},
		{
			name: "Task list",
			in:   `<ul><li><input type="checkbox" checked>done</li><li><input type="checkbox">todo</li></ul>`,
			want: []tg.PageBlockClass{List(
				CheckListItem(true, Plain("done")),
				CheckListItem(false, Plain("todo")),
			)},
		},
		{
			name: "Table",
			in:   "<table><tr><th>A</th><th>B</th></tr><tr><td>1</td><td>2</td></tr></table>",
			want: []tg.PageBlockClass{func() tg.PageBlockClass {
				tbl := Table(Empty(),
					Row(HeaderCell(Plain("A")), HeaderCell(Plain("B"))),
					Row(Cell(Plain("1")), Cell(Plain("2"))),
				)
				tbl.Bordered = true
				return tbl
			}()},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, parseHTML(t, tt.in))
		})
	}
}

func TestParseMarkdown(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   string
		want []tg.PageBlockClass
	}{
		{
			name: "Heading",
			in:   "# Title",
			want: []tg.PageBlockClass{Heading1(Plain("Title"))},
		},
		{
			name: "Bold italic",
			in:   "**bold** and *italic*",
			want: []tg.PageBlockClass{Paragraph(Concat(
				Bold(Plain("bold")), Plain(" and "), Italic(Plain("italic")),
			))},
		},
		{
			name: "Strikethrough",
			in:   "~~gone~~",
			want: []tg.PageBlockClass{Paragraph(Strike(Plain("gone")))},
		},
		{
			name: "Inline code",
			in:   "`code`",
			want: []tg.PageBlockClass{Paragraph(Fixed(Plain("code")))},
		},
		{
			name: "Link",
			in:   "[Go](https://go.dev)",
			want: []tg.PageBlockClass{Paragraph(URL(Plain("Go"), "https://go.dev", 0))},
		},
		{
			name: "Thematic break",
			in:   "a\n\n---\n\nb",
			want: []tg.PageBlockClass{Paragraph(Plain("a")), Divider(), Paragraph(Plain("b"))},
		},
		{
			name: "Fenced code",
			in:   "```go\nx := 1\n```",
			want: []tg.PageBlockClass{Preformatted(Plain("x := 1"), "go")},
		},
		{
			name: "Blockquote",
			in:   "> quoted",
			want: []tg.PageBlockClass{Blockquote(Plain("quoted"), Empty())},
		},
		{
			name: "Unordered list",
			in:   "- one\n- two",
			want: []tg.PageBlockClass{List(ListItem(Plain("one")), ListItem(Plain("two")))},
		},
		{
			name: "Task list",
			in:   "- [x] done\n- [ ] todo",
			want: []tg.PageBlockClass{List(
				CheckListItem(true, Plain("done")),
				CheckListItem(false, Plain("todo")),
			)},
		},
		{
			name: "Embedded subscript HTML",
			in:   "H<sub>2</sub>O",
			want: []tg.PageBlockClass{Paragraph(Concat(
				Plain("H"), Subscript(Plain("2")), Plain("O"),
			))},
		},
		{
			name: "Ordered list",
			in:   "1. one\n2. two",
			want: []tg.PageBlockClass{OrderedList(
				OrderedListItem("", Plain("one")),
				OrderedListItem("", Plain("two")),
			)},
		},
		{
			name: "Multi-paragraph blockquote",
			in:   "> one\n>\n> two",
			want: []tg.PageBlockClass{BlockquoteBlocks(Empty(),
				Paragraph(Plain("one")),
				Paragraph(Plain("two")),
			)},
		},
		{
			name: "Marked",
			in:   "==marked text==",
			want: []tg.PageBlockClass{Paragraph(Marked(Plain("marked text")))},
		},
		{
			name: "Spoiler",
			in:   "||spoiler||",
			want: []tg.PageBlockClass{Paragraph(Spoiler(Plain("spoiler")))},
		},
		{
			name: "Inline math",
			in:   "before $x^2 + y^2$ after",
			want: []tg.PageBlockClass{Paragraph(Concat(
				Plain("before "), Math("x^2 + y^2"), Plain(" after"),
			))},
		},
		{
			name: "Display math",
			in:   "$$E = mc^2$$",
			want: []tg.PageBlockClass{MathBlock("E = mc^2")},
		},
		{
			name: "Fenced math",
			in:   "```math\nE = mc^2\n```",
			want: []tg.PageBlockClass{MathBlock("E = mc^2")},
		},
		{
			name: "Phone link",
			in:   "[call](tel:+123456789)",
			want: []tg.PageBlockClass{Paragraph(Phone(Plain("call"), "+123456789"))},
		},
		{
			name: "User mention link",
			in:   "[me](tg://user?id=123456789)",
			want: []tg.PageBlockClass{Paragraph(MentionName(Plain("me"), 123456789))},
		},
		{
			name: "Custom emoji image",
			in:   "![👍](tg://emoji?id=5368324170671202286)",
			want: []tg.PageBlockClass{Paragraph(CustomEmoji(5368324170671202286, "👍"))},
		},
		{
			name: "Date image",
			in:   "![22:45 tomorrow](tg://time?unix=1647531900&format=wDT)",
			want: []tg.PageBlockClass{Paragraph(Date(Plain("22:45 tomorrow"), 1647531900, DateFlags{
				LongTime:  true,
				LongDate:  true,
				DayOfWeek: true,
			}))},
		},
		{
			name: "Table alignment",
			in:   "| H1 | H2 |\n|:---|:--:|\n| a | b |",
			want: []tg.PageBlockClass{func() tg.PageBlockClass {
				h2 := HeaderCell(Plain("H2"))
				h2.AlignCenter = true
				b := Cell(Plain("b"))
				b.AlignCenter = true
				tbl := Table(Empty(),
					Row(HeaderCell(Plain("H1")), h2),
					Row(Cell(Plain("a")), b),
				)
				tbl.Bordered = true
				return tbl
			}()},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, parseMarkdown(t, tt.in))
		})
	}
}

func TestMessage(t *testing.T) {
	msg := New(Heading1(Plain("Title"))).
		Block(Paragraph(Plain("body"))).
		RTL().
		NoAutoLink().
		Input()

	require.True(t, msg.Rtl)
	require.True(t, msg.Noautolink)
	require.Equal(t, []tg.PageBlockClass{
		Heading1(Plain("Title")),
		Paragraph(Plain("body")),
	}, msg.Blocks)
}

func TestSourceConstructors(t *testing.T) {
	h := Rich().RTL().HTML("<p>hi</p>")
	require.Equal(t, "<p>hi</p>", h.HTML)
	require.True(t, h.Rtl)

	m := Markdown("# hi")
	require.Equal(t, "# hi", m.Markdown)
}
