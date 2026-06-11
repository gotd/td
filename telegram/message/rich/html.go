package rich

import (
	"io"
	"strconv"
	"strings"

	"github.com/go-faster/errors"
	"golang.org/x/net/html"

	"github.com/gotd/td/tg"
)

// ParseHTML parses HTML from r into rich message blocks.
//
// It is a best-effort local renderer that covers the text-expressible subset of
// rich content: paragraphs, headings (h1-h6), dividers (hr), block quotes,
// preformatted/code blocks, ordered and unordered lists (including task lists),
// tables, and inline formatting (bold, italic, underline, strikethrough,
// fixed-width, spoiler, subscript, superscript, marked, links, anchors, custom
// emoji and inline math). Unknown elements are unwrapped to their content.
//
// For full server-side fidelity (media, maps, footnotes and every documented
// tag) send the HTML via [HTML] instead, which lets Telegram parse it.
func ParseHTML(r io.Reader) ([]tg.PageBlockClass, error) {
	root, err := html.Parse(r)
	if err != nil {
		return nil, errors.Wrap(err, "parse html")
	}
	body := findBody(root)
	if body == nil {
		return nil, nil
	}
	return htmlBlocks(body), nil
}

func findBody(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "body" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if b := findBody(c); b != nil {
			return b
		}
	}
	return nil
}

// htmlBlockTags is the set of element names handled as block-level.
var htmlBlockTags = map[string]bool{
	"p": true, "div": true, "section": true, "article": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"hr": true, "blockquote": true, "pre": true,
	"ul": true, "ol": true, "table": true, "details": true,
}

// htmlBlocks renders the block-level children of n, grouping runs of inline
// content into paragraphs.
func htmlBlocks(n *html.Node) []tg.PageBlockClass {
	var (
		blocks []tg.PageBlockClass
		inline []tg.RichTextClass
	)
	flush := func() {
		if t, ok := trimInline(inline); ok {
			blocks = append(blocks, Paragraph(t))
		}
		inline = nil
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && htmlBlockTags[c.Data] {
			flush()
			blocks = append(blocks, htmlBlock(c)...)
			continue
		}
		inline = append(inline, htmlInline(c)...)
	}
	flush()
	return blocks
}

// htmlBlock renders a single block-level element.
func htmlBlock(n *html.Node) []tg.PageBlockClass {
	switch n.Data {
	case "div", "section", "article":
		return htmlBlocks(n)
	case "p":
		t, ok := trimInline(htmlInline(n))
		if !ok {
			return nil
		}
		return []tg.PageBlockClass{Paragraph(t)}
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level, _ := strconv.Atoi(n.Data[1:])
		return []tg.PageBlockClass{Heading(level, join(htmlInline(n)))}
	case "hr":
		return []tg.PageBlockClass{Divider()}
	case "pre":
		lang, text := htmlPre(n)
		return []tg.PageBlockClass{Preformatted(Plain(text), lang)}
	case "blockquote":
		inner := htmlBlocks(n)
		if len(inner) == 1 {
			if p, ok := inner[0].(*tg.PageBlockParagraph); ok {
				return []tg.PageBlockClass{Blockquote(p.Text, Empty())}
			}
		}
		return []tg.PageBlockClass{BlockquoteBlocks(Empty(), inner...)}
	case "ul":
		return []tg.PageBlockClass{List(htmlListItems(n)...)}
	case "ol":
		return []tg.PageBlockClass{OrderedList(htmlOrderedItems(n)...)}
	case "table":
		return []tg.PageBlockClass{htmlTable(n)}
	case "details":
		return []tg.PageBlockClass{htmlDetails(n)}
	default:
		return htmlBlocks(n)
	}
}

// htmlPre extracts the language (from a nested <code class="language-...">) and
// text content of a <pre> block.
func htmlPre(n *html.Node) (lang, text string) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "code" {
			lang = strings.TrimPrefix(attr(c, "class"), "language-")
		}
	}
	return lang, textContent(n)
}

func htmlListItems(n *html.Node) []tg.PageListItemClass {
	var items []tg.PageListItemClass
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode || c.Data != "li" {
			continue
		}
		checkbox, checked, body := htmlListItem(c)
		item := ListItem(body)
		if checkbox {
			item.Checkbox = true
			item.Checked = checked
		}
		items = append(items, item)
	}
	return items
}

func htmlOrderedItems(n *html.Node) []tg.PageListOrderedItemClass {
	var items []tg.PageListOrderedItemClass
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode || c.Data != "li" {
			continue
		}
		checkbox, checked, body := htmlListItem(c)
		item := OrderedListItem("", body)
		if checkbox {
			item.Checkbox = true
			item.Checked = checked
		}
		items = append(items, item)
	}
	return items
}

// htmlListItem renders a list item, detecting a leading task-list checkbox
// (<input type="checkbox">).
func htmlListItem(n *html.Node) (checkbox, checked bool, body tg.RichTextClass) {
	var inline []tg.RichTextClass
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "input" && attr(c, "type") == "checkbox" {
			checkbox = true
			checked = hasAttr(c, "checked")
			continue
		}
		inline = append(inline, htmlInline(c)...)
	}
	return checkbox, checked, join(trimZero(inline))
}

func htmlTable(n *html.Node) *tg.PageBlockTable {
	var rows []tg.PageTableRow
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Type != html.ElementNode {
				continue
			}
			switch c.Data {
			case "thead", "tbody", "tfoot":
				walk(c)
			case "tr":
				rows = append(rows, htmlTableRow(c))
			}
		}
	}
	walk(n)
	t := Table(Empty(), rows...)
	t.Bordered = true
	return t
}

func htmlTableRow(n *html.Node) tg.PageTableRow {
	var cells []tg.PageTableCell
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode || (c.Data != "td" && c.Data != "th") {
			continue
		}
		cell := Cell(join(htmlInline(c)))
		if c.Data == "th" {
			cell.Header = true
		}
		if v, err := strconv.Atoi(attr(c, "colspan")); err == nil && v > 0 {
			cell.Colspan = v
		}
		if v, err := strconv.Atoi(attr(c, "rowspan")); err == nil && v > 0 {
			cell.Rowspan = v
		}
		cells = append(cells, cell)
	}
	return Row(cells...)
}

func htmlDetails(n *html.Node) *tg.PageBlockDetails {
	var (
		title  tg.RichTextClass = Empty()
		blocks []tg.PageBlockClass
		inline []tg.RichTextClass
	)
	flush := func() {
		if t, ok := trimInline(inline); ok {
			blocks = append(blocks, Paragraph(t))
		}
		inline = nil
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "summary" {
			title = join(htmlInline(c))
			continue
		}
		if c.Type == html.ElementNode && htmlBlockTags[c.Data] {
			flush()
			blocks = append(blocks, htmlBlock(c)...)
			continue
		}
		inline = append(inline, htmlInline(c)...)
	}
	flush()
	return Details(hasAttr(n, "open"), title, blocks...)
}

// htmlInline renders a node and its descendants as inline rich text.
func htmlInline(n *html.Node) []tg.RichTextClass {
	switch n.Type {
	case html.TextNode:
		s := collapseWS(n.Data)
		if s == "" {
			return nil
		}
		return []tg.RichTextClass{Plain(s)}
	case html.ElementNode:
		return htmlInlineElement(n)
	default:
		return nil
	}
}

func htmlInlineElement(n *html.Node) []tg.RichTextClass {
	children := func() []tg.RichTextClass {
		var out []tg.RichTextClass
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			out = append(out, htmlInline(c)...)
		}
		return out
	}

	switch n.Data {
	case "b", "strong":
		return []tg.RichTextClass{Bold(children()...)}
	case "i", "em":
		return []tg.RichTextClass{Italic(children()...)}
	case "u", "ins":
		return []tg.RichTextClass{Underline(children()...)}
	case "s", "strike", "del":
		return []tg.RichTextClass{Strike(children()...)}
	case "code", "tt", "kbd", "samp":
		return []tg.RichTextClass{Fixed(children()...)}
	case "sub":
		return []tg.RichTextClass{Subscript(children()...)}
	case "sup":
		return []tg.RichTextClass{Superscript(children()...)}
	case "mark":
		return []tg.RichTextClass{Marked(children()...)}
	case "tg-spoiler":
		return []tg.RichTextClass{Spoiler(children()...)}
	case "span":
		if strings.Contains(attr(n, "class"), "tg-spoiler") {
			return []tg.RichTextClass{Spoiler(children()...)}
		}
		return children()
	case "tg-emoji":
		id, err := strconv.ParseInt(attr(n, "emoji-id"), 10, 64)
		if err != nil {
			return children()
		}
		return []tg.RichTextClass{CustomEmoji(id, textContent(n))}
	case "tg-math", "math":
		return []tg.RichTextClass{Math(textContent(n))}
	case "br":
		return []tg.RichTextClass{Plain("\n")}
	case "a":
		return []tg.RichTextClass{htmlAnchor(n, children())}
	default:
		return children()
	}
}

func htmlAnchor(n *html.Node, children []tg.RichTextClass) tg.RichTextClass {
	text := join(children)
	if name := attr(n, "name"); name != "" {
		return Anchor(text, name)
	}
	href := attr(n, "href")
	switch {
	case href == "":
		return text
	case strings.HasPrefix(href, "#"):
		return AnchorLink(text, strings.TrimPrefix(href, "#"))
	case strings.HasPrefix(href, "mailto:"):
		return Email(text, strings.TrimPrefix(href, "mailto:"))
	default:
		return URL(text, href, 0)
	}
}

// textContent returns the concatenated text of a node and its descendants.
func textContent(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			sb.WriteString(node.Data)
			return
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return sb.String()
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func hasAttr(n *html.Node, key string) bool {
	for _, a := range n.Attr {
		if a.Key == key {
			return true
		}
	}
	return false
}
