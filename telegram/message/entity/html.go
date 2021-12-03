package entity

import (
	"io"
	"strings"

	"github.com/go-faster/errors"
	"golang.org/x/net/html"

	"github.com/gotd/td/tg"
)

type htmlParser struct {
	tokenizer *html.Tokenizer
	builder   *Builder
	offset    int
	stack     stack
	attr      map[string]string
	opts      HTMLOptions
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

func (p *htmlParser) startTag() error {
	const pre = "pre"

	var e stackElem
	tn, hasAttr := p.tokenizer.TagName()
	e.tag = string(tn)
	if hasAttr {
		p.fillAttrs()
	}

	e.offset = p.offset
	e.utf8offset = p.builder.message.Len()
	// See https://core.telegram.org/bots/api#html-style.
	switch e.tag {
	case "b", "strong":
		e.format = Bold()
	case "i", "em":
		e.format = Italic()
	case "u", "ins":
		e.format = Underline()
	case "s", "strike", "del":
		e.format = Strike()
	case "a":
		href := p.attr["href"]
		if href == "" {
			e.tryURL = true
			break
		}

		f, err := getURLFormatter(href, p.opts.UserResolver)
		if err != nil {
			f = nil
		}
		e.format = f
	case "code":
		e.format = Code()
		if len(p.stack) < 1 {
			break
		}

		// BotAPI docs says:
		// > Use nested <pre> and <code> tags, to define programming language for <pre> entity.
		last := &p.stack[len(p.stack)-1]
		if last.tag != pre {
			break
		}

		const langPrefix = "language-"
		if lang := p.attr["class"]; strings.HasPrefix(lang, langPrefix) {
			// Set language parameter.
			last.format = Pre(lang[len(langPrefix):])
		}
	case pre:
		e.format = Pre("")
	}

	p.stack.push(e)
	return nil
}

func (p *htmlParser) endTag(checkName bool) error {
	tn, _ := p.tokenizer.TagName()

	s, ok := p.stack.pop()
	switch  {
	case !ok:
		return errors.Errorf("unexpected end tag %q", tn)
	case checkName && s.tag != string(tn) :
		return errors.Errorf("expected tag %q, got %q", s.tag, tn)
	}

	// Compute UTF-16 length of entity.
	length := ComputeLength(p.builder.message.String()) - s.offset
	utf8Length := p.builder.message.Len() - s.utf8offset
	if s.tag == "a" && s.tryURL {
		msg := p.builder.message.String()[s.utf8offset : s.utf8offset+utf8Length]
		if f, err := getURLFormatter(msg, p.opts.UserResolver); err == nil {
			s.format = f
		}
	}

	// Do not add empty entities.
	if length == 0 || s.format == nil {
		return nil
	}

	u8 := utf8entity{
		offset: s.utf8offset,
		length: utf8Length,
	}
	p.builder.appendEntities(s.offset, length, u8, s.format)
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
			p.builder.message.Write(text)
			p.offset += ComputeLength(string(text))
		case html.StartTagToken:
			if err := p.startTag(); err != nil {
				return err
			}
		case html.EndTagToken:
			if err := p.endTag(true); err != nil {
				return err
			}
		case html.CommentToken:
			raw := p.tokenizer.Raw()
			if len(raw) >= 3 && string(raw[:2]) == "</" && raw[len(raw)-1] == '>' {
				if err := p.endTag(false); err != nil {
					return err
				}
			}
		}
	}
}

// HTMLOptions is options of HTML.
type HTMLOptions struct {
	// UserResolver is used to resolve user by ID during formatting. May be nil.
	//
	// If userResolver is nil, formatter will create tg.InputUser using only ID.
	// Notice that it's okay for bots, but not for users.
	UserResolver UserResolver
	// DisableTelegramEscape disable Telegram BotAPI escaping and uses default
	// golang.org/x/net/html escape.
	DisableTelegramEscape bool
}

func (o *HTMLOptions) setDefaults() {
	if o.UserResolver == nil {
		o.UserResolver = func(id int64) (tg.InputUserClass, error) {
			return &tg.InputUser{
				UserID: id,
			}, nil
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
func HTML(r io.Reader, b *Builder, opts HTMLOptions) error {
	opts.setDefaults()
	p := htmlParser{
		tokenizer: html.NewTokenizer(r),
		builder:   b,
		attr:      map[string]string{},
		opts:      opts,
	}
	return p.parse()
}
