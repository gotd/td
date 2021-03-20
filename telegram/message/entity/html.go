package entity

import (
	"io"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/xerrors"
)

type stackElem struct {
	offset int
	tag    string
	format Formatter
}

// HTML parses given input from reader and adds parsed entities to given builder.
//
// See https://core.telegram.org/bots/api#html-style.
func HTML(r io.Reader, b *Builder) error {
	tokenizer := html.NewTokenizer(r)

	var stack []stackElem
	attr := map[string]string{}
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			if err := tokenizer.Err(); !xerrors.Is(err, io.EOF) {
				return err
			}
			return nil
		case html.TextToken:
			b.message.Write(tokenizer.Text())
		case html.StartTagToken:
			var e stackElem
			tn, hasAttr := tokenizer.TagName()
			e.tag = string(tn)

			if hasAttr {
				// Clear old attrs.
				for k := range attr {
					delete(attr, k)
				}

				// Fill with new attributes.
				for {
					key, value, ok := tokenizer.TagAttr()
					attr[string(key)] = string(value)
					if !ok {
						break
					}
				}
			}

			e.offset = b.message.Len()
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
				href, ok := attr["href"]
				if !ok {
					return xerrors.Errorf("tag %q must have attribute href", e.tag)
				}

				u, err := url.Parse(href)
				if err != nil {
					return xerrors.Errorf("href must be a valid URL, got %q", href)
				}

				if u.Scheme == "tg" && u.Host == "user" {
					id, err := strconv.Atoi(u.Query().Get("id"))
					if err != nil {
						return xerrors.Errorf("invalid user ID %q: %w", id, err)
					}

					e.format = MentionName(id)
				} else {
					e.format = TextURL(href)
				}
			case "code":
				e.format = Code()

				if len(stack) > 0 && stack[0].tag == "pre" {
					lang, ok := attr["class"]
					if ok {
						e.format = Pre(strings.TrimPrefix(lang, "language-"))
					}
				}
			case "pre":
				e.format = Code()
			default:
				return xerrors.Errorf("unknown tag name %q", e.tag)
			}

			stack = append(stack, e)
		case html.EndTagToken:
			tn, _ := tokenizer.TagName()
			if len(stack) == 0 {
				return xerrors.Errorf("unexpected end tag %q", string(tn))
			}

			var s stackElem
			// Pop from SliceTricks.
			s, stack = stack[len(stack)-1], stack[:len(stack)-1]
			if s.tag != string(tn) {
				return xerrors.Errorf("expected tag %q, got %q", s.tag, string(tn))
			}

			length := ComputeLength(b.message.String())
			b.entities = append(b.entities, s.format(s.offset, length-s.offset))
		}
	}
}
