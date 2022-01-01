package html

import (
	"io"
	"strings"

	"github.com/go-faster/errors"
	"golang.org/x/net/html"

	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
)

type htmlParser struct {
	tokenizer *html.Tokenizer
	builder   *entity.Builder
	stack     stack
	attr      map[string]string
	opts      Options
}

func (p *htmlParser) fillAttrs() {
	// Clear old attrs.
	for k := range p.attr {
		delete(p.attr, k)
	}

	// Fill with new attributes.
	for {
		key, value, ok := p.tokenizer.TagAttr()
		p.attr[string(key)] = string(value)
		if !ok {
			break
		}
	}
}

const (
	pre       = "pre"
	code      = "code"
	em        = "em"
	ins       = "ins"
	strike    = "strike"
	del       = "del"
	strong    = "strong"
	span      = "span"
	tgSpoiler = "tg-spoiler"
)

func (p *htmlParser) tag(tn []byte) string {
	// Here we intern some well-known tags.
	switch string(tn) {
	case "b":
		return "b"
	case strong:
		return strong
	case "i":
		return "i"
	case em:
		return em
	case "u":
		return "u"
	case ins:
		return ins
	case "s":
		return "s"
	case strike:
		return strike
	case del:
		return del
	case "a":
		return "a"
	case pre:
		return pre
	case code:
		return code
	case span:
		return span
	case tgSpoiler:
		return tgSpoiler
	default:
		return string(tn)
	}
}

func (p *htmlParser) startTag() error {
	var e stackElem
	tn, hasAttr := p.tokenizer.TagName()
	e.tag = p.tag(tn)
	if hasAttr {
		p.fillAttrs()
	}

	e.token = p.builder.Token()
	// See https://core.telegram.org/bots/api#html-style.
	switch e.tag {
	case "b", strong:
		e.format = entity.Bold()
	case "i", em:
		e.format = entity.Italic()
	case "u", ins:
		e.format = entity.Underline()
	case "s", strike, del:
		e.format = entity.Strike()
	case "a":
		e.attr = p.attr["href"]
		if e.attr == "" {
			break
		}

		f, err := getURLFormatter(e.attr, p.opts.UserResolver)
		if err != nil {
			f = nil
		}
		e.format = f
	case code:
		const langPrefix = "language-"

		e.format = entity.Code()
		e.attr = strings.TrimPrefix(p.attr["class"], langPrefix)
		if len(p.stack) < 1 {
			break
		}

		// BotAPI docs says:
		// > Use nested <pre> and <code> tags, to define programming language for <pre> entity.
		last := &p.stack[len(p.stack)-1]
		if last.tag != pre {
			break
		}

		if lang := e.attr; lang != "" {
			// Set language parameter.
			last.format = entity.Pre(lang)
		}
	case pre:
		e.format = entity.Pre("")
		if len(p.stack) < 1 {
			break
		}

		last := &p.stack[len(p.stack)-1]
		if last.tag != code {
			break
		}

		if lang := last.attr; lang != "" {
			// Set language parameter.
			e.format = entity.Pre(lang)
		}
	case span:
		if p.attr["class"] == "tg-spoiler" {
			e.format = entity.Spoiler()
		}
	case tgSpoiler:
		e.format = entity.Spoiler()
	}

	p.stack.push(e)
	return nil
}

func (p *htmlParser) endTag(checkName bool) error {
	tn, _ := p.tokenizer.TagName()

	s, ok := p.stack.pop()
	switch {
	case !ok:
		return errors.Errorf("unexpected end tag %q", tn)
	case checkName && s.tag != string(tn):
		return errors.Errorf("expected tag %q, got %q", s.tag, tn)
	}

	// Compute UTF-16 length of entity.
	length := s.token.UTF16Length(p.builder)

	switch s.tag {
	case "a":
		// TDLib tries to parse link from <a> body, so we should too.
		if s.attr == "" {
			msg := s.token.Text(p.builder)
			if f, err := getURLFormatter(msg, p.opts.UserResolver); err == nil {
				s.format = f
			}
		}
	case "code":
		l, ok := p.builder.LastEntity()
		if !ok {
			break
		}
		last, ok := l.(*tg.MessageEntityPre)
		if !ok {
			break
		}
		// Do not add Code entity, if last entity is Pre with same offset.
		if last.GetOffset() == s.token.UTF16Offset() && last.GetLength() == length {
			return nil
		}
	}
	// Do not add empty entities.
	if length == 0 || s.format == nil {
		return nil
	}

	s.token.Apply(p.builder, s.format)
	return nil
}

func (p *htmlParser) parse() error {
	for {
		tt := p.tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			if err := p.tokenizer.Err(); !errors.Is(err, io.EOF) {
				return err
			}
			return nil
		case html.TextToken:
			var text []byte
			if p.opts.DisableTelegramEscape {
				text = p.tokenizer.Text()
			} else {
				text = telegramUnescape(p.tokenizer.Raw())
			}
			_, _ = p.builder.Write(text)
		case html.StartTagToken:
			if err := p.startTag(); err != nil {
				return err
			}
		case html.EndTagToken:
			if err := p.endTag(true); err != nil {
				return err
			}
		case html.CommentToken:
			// html.Tokenizer returns comment token for empty closing tags.
			raw := p.tokenizer.Raw()
			if len(raw) >= 3 && string(raw[:2]) == "</" && raw[len(raw)-1] == '>' {
				if err := p.endTag(false); err != nil {
					return err
				}
			}
		}
	}
}

// HTML parses given input from reader and adds parsed entities to given builder.
// Notice that this parser ignores unsupported tags.
//
// Parameter userResolver is used to resolve user by ID during formatting. May be nil.
// If userResolver is nil, formatter will create tg.InputUser using only ID.
// Notice that it's okay for bots, but not for users.
//
// See https://core.telegram.org/bots/api#html-style.
func HTML(r io.Reader, b *entity.Builder, opts Options) error {
	opts.setDefaults()
	p := htmlParser{
		tokenizer: html.NewTokenizer(r),
		builder:   b,
		attr:      map[string]string{},
		opts:      opts,
	}

	if err := p.parse(); err != nil {
		return errors.Wrap(err, "parse")
	}
	b.ShrinkPreCode()
	return nil
}
