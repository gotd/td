package markdown

import (
	"bytes"
	"net/url"
	"strconv"

	"github.com/go-faster/errors"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"

	"github.com/gotd/td/telegram/message/entity"
)

// newMarkdownParser builds a goldmark parser configured to produce only the
// constructs that map onto Telegram message entities.
func newMarkdownParser() parser.Parser {
	return parser.NewParser(
		parser.WithBlockParsers(
			util.Prioritized(parser.NewFencedCodeBlockParser(), 700),
			util.Prioritized(parser.NewBlockquoteParser(), 800),
			util.Prioritized(parser.NewParagraphParser(), 1000),
		),
		parser.WithInlineParsers(
			util.Prioritized(parser.NewCodeSpanParser(), 100),
			util.Prioritized(parser.NewLinkParser(), 200),
			util.Prioritized(parser.NewEmphasisParser(), 500),
			util.Prioritized(extension.NewStrikethroughParser(), 500),
		),
	)
}

// mdParser is safe for concurrent use.
var mdParser = newMarkdownParser()

// renderer converts a goldmark AST into entity.Builder writes.
type renderer struct {
	builder *entity.Builder
	source  []byte
	opts    Options
}

func (r *renderer) renderDocument(doc ast.Node) error {
	return r.renderBlocks(doc)
}

// renderBlocks renders block-level children, separating them with a blank line.
func (r *renderer) renderBlocks(n ast.Node) error {
	first := true
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if !first {
			_, _ = r.builder.WriteString("\n\n")
		}
		first = false
		if err := r.renderBlock(c); err != nil {
			return err
		}
	}
	return nil
}

func (r *renderer) renderBlock(n ast.Node) error {
	switch n := n.(type) {
	case *ast.Paragraph, *ast.TextBlock:
		return r.renderInlines(n)
	case *ast.Blockquote:
		token := r.builder.Token()
		if err := r.renderBlocks(n); err != nil {
			return err
		}
		if token.UTF16Length(r.builder) > 0 {
			token.Apply(r.builder, entity.Blockquote(false))
		}
		return nil
	case *ast.FencedCodeBlock:
		token := r.builder.Token()
		var code []byte
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			seg := lines.At(i)
			code = append(code, seg.Value(r.source)...)
		}
		_, _ = r.builder.Write(bytes.TrimRight(code, "\n"))
		if token.UTF16Length(r.builder) > 0 {
			token.Apply(r.builder, entity.Pre(string(n.Language(r.source))))
		}
		return nil
	default:
		// Unknown block: render its children, if any.
		return r.renderBlocks(n)
	}
}

func (r *renderer) renderInlines(n ast.Node) error {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if err := r.renderInline(c); err != nil {
			return err
		}
	}
	return nil
}

func (r *renderer) renderInline(n ast.Node) error {
	switch n := n.(type) {
	case *ast.Text:
		_, _ = r.builder.Write(util.UnescapePunctuations(n.Segment.Value(r.source)))
		if n.SoftLineBreak() || n.HardLineBreak() {
			_ = r.builder.WriteByte('\n')
		}
		return nil
	case *ast.String:
		_, _ = r.builder.Write(n.Value)
		return nil
	case *ast.CodeSpan:
		return r.styled(n, entity.Code())
	case *ast.Emphasis:
		// Level 1 (*text*, _text_) is italic, level 2 (**text**, __text__) is bold.
		if n.Level >= 2 {
			return r.styled(n, entity.Bold())
		}
		return r.styled(n, entity.Italic())
	case *east.Strikethrough:
		return r.styled(n, entity.Strike())
	case *ast.Link:
		return r.renderLink(n, string(n.Destination), false)
	case *ast.Image:
		return r.renderLink(n, string(n.Destination), true)
	default:
		return r.renderInlines(n)
	}
}

// styled renders inline children and wraps them with the given formatter.
func (r *renderer) styled(n ast.Node, f entity.Formatter) error {
	token := r.builder.Token()
	switch n.(type) {
	case *ast.CodeSpan:
		r.writeRaw(n)
	default:
		if err := r.renderInlines(n); err != nil {
			return err
		}
	}
	if token.UTF16Length(r.builder) > 0 {
		token.Apply(r.builder, f)
	}
	return nil
}

func (r *renderer) renderLink(n ast.Node, dest string, emoji bool) error {
	token := r.builder.Token()
	if err := r.renderInlines(n); err != nil {
		return err
	}
	if token.UTF16Length(r.builder) == 0 {
		return nil
	}
	f, err := r.urlFormatter(dest, emoji)
	if err != nil {
		return err
	}
	if f == nil {
		return nil
	}
	token.Apply(r.builder, f)
	return nil
}

// writeRaw writes the plain text of descendant text nodes without formatting.
func (r *renderer) writeRaw(n ast.Node) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c := c.(type) {
		case *ast.Text:
			_, _ = r.builder.Write(c.Segment.Value(r.source))
		case *ast.String:
			_, _ = r.builder.Write(c.Value)
		default:
			r.writeRaw(c)
		}
	}
}

func (r *renderer) urlFormatter(rawURL string, emoji bool) (entity.Formatter, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, errors.Wrapf(err, "parse URL %q", rawURL)
	}

	switch {
	case u.Scheme == "tg" && u.Host == "emoji":
		// Custom emoji: ![emoji](tg://emoji?id=N).
		id, err := strconv.ParseInt(u.Query().Get("id"), 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "parse custom emoji ID")
		}
		return entity.CustomEmoji(id), nil
	case emoji:
		// An image without a custom emoji URL has no entity equivalent; keep
		// its alt text as plain text.
		return nil, nil
	case u.Scheme == "tg" && u.Host == "user":
		// Inline mention: [name](tg://user?id=N).
		id, err := strconv.ParseInt(u.Query().Get("id"), 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "parse user ID")
		}
		user, err := r.opts.UserResolver(id)
		if err != nil {
			return nil, errors.Wrapf(err, "resolve user %d", id)
		}
		return entity.MentionName(user), nil
	case rawURL == "":
		return nil, nil
	default:
		return entity.TextURL(rawURL), nil
	}
}
