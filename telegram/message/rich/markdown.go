package rich

import (
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-faster/errors"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"golang.org/x/net/html"

	"github.com/gotd/td/tg"
)

// mdRichParser parses GitHub Flavored Markdown (tables, task lists,
// strikethrough, autolinks) plus raw HTML and the rich-only inline syntax
// ==marked==, ||spoiler|| and $math$.
var mdRichParser = newMarkdownParser()

func newMarkdownParser() parser.Parser {
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))
	md.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(&mathParser{}, 150),
		util.Prioritized(newPairParser('='), 500),
		util.Prioritized(newPairParser('|'), 500),
	))
	return md.Parser()
}

// ParseMarkdown parses Markdown from r into rich message blocks.
//
// It is a best-effort local renderer following Rich Markdown (GitHub Flavored
// Markdown plus Telegram extensions):
//
//   - Blocks: headings, paragraphs, thematic breaks, block quotes,
//     fenced/indented code, ordered and unordered lists (including task lists),
//     tables (with column alignment), and block math via $$...$$ or ```math.
//   - Inline: **bold**, *italic*, ~~strikethrough~~, `code`, ==marked==,
//     ||spoiler|| and $math$.
//   - Links: [t](url), [t](mailto:..), [t](tel:..) and [t](tg://user?id=N);
//     images ![alt](tg://emoji?id=N) and ![alt](tg://time?unix=N&format=F).
//   - Any other rich inline styles (underline, subscript, superscript, anchors,
//     ...) are recognized when written as embedded HTML tags (<u>, <sub>,
//     <sup>, <a name>, ...), as Rich Markdown allows.
//
// Media images referencing a URL (![](photo.jpg)) and footnotes cannot be
// resolved locally; for full server-side fidelity send the Markdown via
// [Markdown] instead, which lets Telegram parse it.
func ParseMarkdown(r io.Reader) ([]tg.PageBlockClass, error) {
	source, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "read")
	}
	doc := mdRichParser.Parse(text.NewReader(source))
	return mdBlocks(doc, source), nil
}

// mdBlocks renders the block-level children of n.
func mdBlocks(n ast.Node, source []byte) []tg.PageBlockClass {
	var blocks []tg.PageBlockClass
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		blocks = append(blocks, mdBlock(c, source)...)
	}
	return blocks
}

func mdBlock(n ast.Node, source []byte) []tg.PageBlockClass {
	switch n := n.(type) {
	case *ast.Heading:
		return []tg.PageBlockClass{Heading(n.Level, join(mdInline(n, source)))}
	case *ast.Paragraph, *ast.TextBlock:
		// A paragraph holding only a $$...$$ expression is a block-level formula.
		if src, ok := soleDisplayMath(n); ok {
			return []tg.PageBlockClass{MathBlock(src)}
		}
		if t, ok := trimInline(mdInline(n, source)); ok {
			return []tg.PageBlockClass{Paragraph(t)}
		}
		return nil
	case *ast.ThematicBreak:
		return []tg.PageBlockClass{Divider()}
	case *ast.Blockquote:
		inner := mdBlocks(n, source)
		if len(inner) == 1 {
			if p, ok := inner[0].(*tg.PageBlockParagraph); ok {
				return []tg.PageBlockClass{Blockquote(p.Text, Empty())}
			}
		}
		return []tg.PageBlockClass{BlockquoteBlocks(Empty(), inner...)}
	case *ast.FencedCodeBlock:
		// A ```math fenced block is a block-level formula.
		if lang := string(n.Language(source)); lang == "math" {
			return []tg.PageBlockClass{MathBlock(mdCodeText(n, source))}
		}
		return []tg.PageBlockClass{Preformatted(Plain(mdCodeText(n, source)), string(n.Language(source)))}
	case *ast.CodeBlock:
		return []tg.PageBlockClass{Preformatted(Plain(mdCodeText(n, source)), "")}
	case *ast.List:
		return []tg.PageBlockClass{mdList(n, source)}
	case *east.Table:
		return []tg.PageBlockClass{mdTable(n, source)}
	case *ast.HTMLBlock:
		blocks, err := ParseHTML(strings.NewReader(mdRawText(n, source)))
		if err != nil {
			return nil
		}
		return blocks
	default:
		return mdBlocks(n, source)
	}
}

func mdList(n *ast.List, source []byte) tg.PageBlockClass {
	if n.IsOrdered() {
		var items []tg.PageListOrderedItemClass
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			item, ok := c.(*ast.ListItem)
			if !ok {
				continue
			}
			checkbox, checked := mdTaskCheckbox(item)
			it := OrderedListItem("", mdItemContent(item, source))
			if checkbox {
				it.Checkbox = true
				it.Checked = checked
			}
			items = append(items, it)
		}
		return OrderedList(items...)
	}

	var items []tg.PageListItemClass
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		item, ok := c.(*ast.ListItem)
		if !ok {
			continue
		}
		checkbox, checked := mdTaskCheckbox(item)
		it := ListItem(mdItemContent(item, source))
		if checkbox {
			it.Checkbox = true
			it.Checked = checked
		}
		items = append(items, it)
	}
	return List(items...)
}

// mdItemContent renders a list item as a single inline rich text, joining its
// block children.
func mdItemContent(item *ast.ListItem, source []byte) tg.RichTextClass {
	var texts []tg.RichTextClass
	for c := item.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.(type) {
		case *ast.Paragraph, *ast.TextBlock:
			texts = append(texts, mdInline(c, source)...)
		}
	}
	return join(trimZero(texts))
}

func mdTaskCheckbox(item *ast.ListItem) (checkbox, checked bool) {
	fc := item.FirstChild()
	if fc == nil {
		return false, false
	}
	if cb, ok := fc.FirstChild().(*east.TaskCheckBox); ok {
		return true, cb.IsChecked
	}
	return false, false
}

func mdTable(n *east.Table, source []byte) tg.PageBlockClass {
	var rows []tg.PageTableRow
	for r := n.FirstChild(); r != nil; r = r.NextSibling() {
		switch r := r.(type) {
		case *east.TableHeader:
			rows = append(rows, mdTableRow(r, source, true))
		case *east.TableRow:
			rows = append(rows, mdTableRow(r, source, false))
		}
	}
	t := Table(Empty(), rows...)
	t.Bordered = true
	return t
}

func mdTableRow(n ast.Node, source []byte, header bool) tg.PageTableRow {
	var cells []tg.PageTableCell
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		tc, ok := c.(*east.TableCell)
		if !ok {
			continue
		}
		cell := Cell(join(mdInline(tc, source)))
		cell.Header = header
		switch tc.Alignment {
		case east.AlignCenter:
			cell.AlignCenter = true
		case east.AlignRight:
			cell.AlignRight = true
		}
		cells = append(cells, cell)
	}
	return Row(cells...)
}

// mdInline renders the inline children of n into rich text, tracking embedded
// raw HTML tags with a stack so paired tags such as <sub>..</sub> wrap their
// content.
func mdInline(n ast.Node, source []byte) []tg.RichTextClass {
	type frame struct {
		name string
		wrap func([]tg.RichTextClass) tg.RichTextClass
		acc  []tg.RichTextClass
	}
	stack := []*frame{{}}
	top := func() *frame { return stack[len(stack)-1] }
	emit := func(ts ...tg.RichTextClass) { top().acc = append(top().acc, ts...) }
	pop := func() {
		f := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		top().acc = append(top().acc, f.wrap(f.acc))
	}

	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		raw, ok := c.(*ast.RawHTML)
		if !ok {
			emit(mdInlineNode(c, source)...)
			continue
		}
		name, closing, selfClose, attrs := parseTag(mdRawText(raw, source))
		switch {
		case name == "":
			continue
		case closing:
			for i := len(stack) - 1; i >= 1; i-- {
				if stack[i].name == name {
					for len(stack) > i {
						pop()
					}
					break
				}
			}
		case name == "br":
			emit(Plain("\n"))
		default:
			wrap, supported := inlineWrapper(name, attrs)
			if !supported {
				continue
			}
			if selfClose {
				emit(wrap(nil))
				continue
			}
			stack = append(stack, &frame{name: name, wrap: wrap})
		}
	}
	for len(stack) > 1 {
		pop()
	}
	return stack[0].acc
}

func mdInlineNode(c ast.Node, source []byte) []tg.RichTextClass {
	switch c := c.(type) {
	case *ast.Text:
		out := []tg.RichTextClass{Plain(string(util.UnescapePunctuations(c.Segment.Value(source))))}
		if c.SoftLineBreak() || c.HardLineBreak() {
			out = append(out, Plain("\n"))
		}
		return out
	case *ast.String:
		return []tg.RichTextClass{Plain(string(c.Value))}
	case *ast.CodeSpan:
		return []tg.RichTextClass{Fixed(Plain(astText(c, source)))}
	case *ast.Emphasis:
		inner := mdInline(c, source)
		if c.Level >= 2 {
			return []tg.RichTextClass{Bold(inner...)}
		}
		return []tg.RichTextClass{Italic(inner...)}
	case *east.Strikethrough:
		return []tg.RichTextClass{Strike(mdInline(c, source)...)}
	case *markedNode:
		return []tg.RichTextClass{Marked(mdInline(c, source)...)}
	case *spoilerNode:
		return []tg.RichTextClass{Spoiler(mdInline(c, source)...)}
	case *mathNode:
		return []tg.RichTextClass{Math(c.Source)}
	case *ast.Link:
		return []tg.RichTextClass{mdLink(string(c.Destination), mdInline(c, source))}
	case *ast.AutoLink:
		label := Plain(string(c.Label(source)))
		url := string(c.URL(source))
		if c.AutoLinkType == ast.AutoLinkEmail {
			return []tg.RichTextClass{Email(label, url)}
		}
		return []tg.RichTextClass{URL(label, url, 0)}
	case *ast.Image:
		return []tg.RichTextClass{mdImage(string(c.Destination), mdInline(c, source))}
	default:
		return mdInline(c, source)
	}
}

// soleDisplayMath reports whether n's only meaningful inline child is a $$...$$
// display-math node and returns its source.
func soleDisplayMath(n ast.Node) (string, bool) {
	var math *mathNode
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok && t.SoftLineBreak() && t.Segment.Len() == 0 {
			continue
		}
		m, ok := c.(*mathNode)
		if !ok || math != nil {
			return "", false
		}
		math = m
	}
	if math == nil || !math.Display {
		return "", false
	}
	return math.Source, true
}

func mdLink(dest string, inner []tg.RichTextClass) tg.RichTextClass {
	text := join(inner)
	switch {
	case dest == "":
		return text
	case strings.HasPrefix(dest, "#"):
		return AnchorLink(text, strings.TrimPrefix(dest, "#"))
	case strings.HasPrefix(dest, "mailto:"):
		return Email(text, strings.TrimPrefix(dest, "mailto:"))
	case strings.HasPrefix(dest, "tel:"):
		return Phone(text, strings.TrimPrefix(dest, "tel:"))
	}

	if u, err := url.Parse(dest); err == nil && u.Scheme == "tg" && u.Host == "user" {
		if id, err := strconv.ParseInt(u.Query().Get("id"), 10, 64); err == nil {
			return MentionName(text, id)
		}
	}
	return URL(text, dest, 0)
}

// mdImage renders a Markdown image. tg://emoji and tg://time destinations map to
// custom emoji and formatted-date rich text; media destinations cannot be
// resolved locally and fall back to their alt text.
func mdImage(dest string, alt []tg.RichTextClass) tg.RichTextClass {
	u, err := url.Parse(dest)
	if err != nil || u.Scheme != "tg" {
		return join(alt)
	}
	switch u.Host {
	case "emoji":
		id, err := strconv.ParseInt(u.Query().Get("id"), 10, 64)
		if err != nil {
			return join(alt)
		}
		return CustomEmoji(id, plainText(alt))
	case "time":
		unix, err := strconv.Atoi(u.Query().Get("unix"))
		if err != nil || unix <= 0 {
			return join(alt)
		}
		return Date(join(alt), unix, parseDateFlags(u.Query().Get("format")))
	default:
		return join(alt)
	}
}

// parseDateFlags parses a Telegram date format string (as in
// tg://time?format=...) into DateFlags. "r"/"R" select relative time and must
// be the only character; otherwise each character toggles a component.
func parseDateFlags(format string) DateFlags {
	if format == "r" || format == "R" {
		return DateFlags{Relative: true}
	}
	var f DateFlags
	for _, c := range format {
		switch c {
		case 't':
			f.ShortTime = true
		case 'T':
			f.LongTime = true
		case 'd':
			f.ShortDate = true
		case 'D':
			f.LongDate = true
		case 'w', 'W':
			f.DayOfWeek = true
		}
	}
	return f
}

// inlineWrapper maps an inline HTML tag onto a rich text wrapper.
func inlineWrapper(name string, attrs map[string]string) (func([]tg.RichTextClass) tg.RichTextClass, bool) {
	switch name {
	case "b", "strong":
		return func(ts []tg.RichTextClass) tg.RichTextClass { return Bold(ts...) }, true
	case "i", "em":
		return func(ts []tg.RichTextClass) tg.RichTextClass { return Italic(ts...) }, true
	case "u", "ins":
		return func(ts []tg.RichTextClass) tg.RichTextClass { return Underline(ts...) }, true
	case "s", "strike", "del":
		return func(ts []tg.RichTextClass) tg.RichTextClass { return Strike(ts...) }, true
	case "code", "tt", "kbd", "samp":
		return func(ts []tg.RichTextClass) tg.RichTextClass { return Fixed(ts...) }, true
	case "sub":
		return func(ts []tg.RichTextClass) tg.RichTextClass { return Subscript(ts...) }, true
	case "sup":
		return func(ts []tg.RichTextClass) tg.RichTextClass { return Superscript(ts...) }, true
	case "mark":
		return func(ts []tg.RichTextClass) tg.RichTextClass { return Marked(ts...) }, true
	case "tg-spoiler":
		return func(ts []tg.RichTextClass) tg.RichTextClass { return Spoiler(ts...) }, true
	case "span":
		if strings.Contains(attrs["class"], "tg-spoiler") {
			return func(ts []tg.RichTextClass) tg.RichTextClass { return Spoiler(ts...) }, true
		}
		return nil, false
	case "tg-math", "math":
		return func(ts []tg.RichTextClass) tg.RichTextClass { return Math(plainText(ts)) }, true
	case "tg-emoji":
		id, err := strconv.ParseInt(attrs["emoji-id"], 10, 64)
		if err != nil {
			return nil, false
		}
		return func(ts []tg.RichTextClass) tg.RichTextClass { return CustomEmoji(id, plainText(ts)) }, true
	case "a":
		return func(ts []tg.RichTextClass) tg.RichTextClass {
			if name := attrs["name"]; name != "" {
				return Anchor(join(ts), name)
			}
			return mdLink(attrs["href"], ts)
		}, true
	default:
		return nil, false
	}
}

// parseTag parses a single HTML tag, returning its lower-cased name, whether it
// is a closing or self-closing tag, and its attributes.
func parseTag(raw string) (name string, closing, selfClose bool, attrs map[string]string) {
	z := html.NewTokenizer(strings.NewReader(raw))
	attrs = map[string]string{}
	switch z.Next() {
	case html.StartTagToken, html.SelfClosingTagToken, html.EndTagToken:
		tn, hasAttr := z.TagName()
		name = string(tn)
		raw = strings.TrimSpace(raw)
		closing = strings.HasPrefix(raw, "</")
		selfClose = strings.HasSuffix(raw, "/>")
		if hasAttr {
			for {
				k, v, more := z.TagAttr()
				attrs[string(k)] = string(v)
				if !more {
					break
				}
			}
		}
	}
	return name, closing, selfClose, attrs
}

// astText returns the concatenated literal text of an inline AST node.
func astText(n ast.Node, source []byte) string {
	var b strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			b.Write(t.Segment.Value(source))
		} else {
			b.WriteString(astText(c, source))
		}
	}
	return b.String()
}

// mdCodeText returns the raw text of a code block.
func mdCodeText(n ast.Node, source []byte) string {
	var b strings.Builder
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		b.Write(seg.Value(source))
	}
	return strings.TrimRight(b.String(), "\n")
}

// mdRawText returns the raw source of an HTML node (block or inline).
func mdRawText(n ast.Node, source []byte) string {
	var b strings.Builder
	switch n := n.(type) {
	case *ast.RawHTML:
		for i := 0; i < n.Segments.Len(); i++ {
			seg := n.Segments.At(i)
			b.Write(seg.Value(source))
		}
	case *ast.HTMLBlock:
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			seg := lines.At(i)
			b.Write(seg.Value(source))
		}
		if n.HasClosure() {
			b.Write(n.ClosureLine.Value(source))
		}
	}
	return b.String()
}

// plainText extracts the plain text content of rich text nodes (best effort,
// used for math sources and custom emoji alt text).
func plainText(texts []tg.RichTextClass) string {
	var b strings.Builder
	var walk func(tg.RichTextClass)
	walk = func(t tg.RichTextClass) {
		switch t := t.(type) {
		case *tg.TextPlain:
			b.WriteString(t.Text)
		case *tg.TextConcat:
			for _, x := range t.Texts {
				walk(x)
			}
		}
	}
	for _, t := range texts {
		walk(t)
	}
	return b.String()
}
