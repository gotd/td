// Package markdown contains Markdown (CommonMark) styling options.
//
// It parses standard Markdown and converts it into Telegram message entities.
package markdown

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/go-faster/errors"
	"github.com/yuin/goldmark/text"

	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

// Markdown parses given input from reader as Markdown and adds parsed entities
// to the given builder.
//
// Parsing is backed by goldmark and follows CommonMark (plus GFM
// strikethrough). The constructs that map onto Telegram message entities are:
//
//	*text*, _text_       italic
//	**text**, __text__   bold
//	~~text~~             strikethrough
//	`text`               inline code
//	```lang              pre-formatted code block
//	[text](url)          inline URL (tg://user?id=N becomes a mention)
//	![e](tg://emoji?id=N) custom emoji
//	> text               block quotation
//
// Constructs without an entity equivalent (headings, lists, images, etc.) are
// rendered as plain text. The parser is lenient: unmatched markup is emitted as
// plain text instead of failing.
//
// Parameter UserResolver of opts is used to resolve user by ID during formatting.
// May be nil. If UserResolver is nil, formatter will create tg.InputUser using
// only ID. Notice that it's okay for bots, but not for users.
func Markdown(r io.Reader, b *entity.Builder, opts Options) error {
	opts.setDefaults()

	source, err := io.ReadAll(r)
	if err != nil {
		return errors.Wrap(err, "read")
	}

	doc := mdParser.Parse(text.NewReader(source))
	rdr := renderer{
		builder: b,
		source:  source,
		opts:    opts,
	}
	if err := rdr.renderDocument(doc); err != nil {
		return errors.Wrap(err, "render")
	}
	b.ShrinkPreCode()
	return nil
}

// Bytes reads Markdown from given byte slice and returns styling option
// to build styled text block.
func Bytes(resolver func(id int64) (tg.InputUserClass, error), b []byte) styling.StyledTextOption {
	return Reader(resolver, bytes.NewReader(b))
}

// String reads Markdown from given string and returns styling option
// to build styled text block.
func String(resolver func(id int64) (tg.InputUserClass, error), s string) styling.StyledTextOption {
	return Reader(resolver, strings.NewReader(s))
}

// Format formats string using fmt, parses Markdown from formatted string
// and returns styling option to build styled text block.
func Format(resolver func(id int64) (tg.InputUserClass, error), format string, args ...interface{}) styling.StyledTextOption {
	return styling.Custom(func(eb *entity.Builder) error {
		var buf bytes.Buffer
		if _, err := fmt.Fprintf(&buf, format, args...); err != nil {
			return err
		}
		return Markdown(&buf, eb, Options{
			UserResolver: resolver,
		})
	})
}

// Reader reads Markdown from given reader and returns styling option
// to build styled text block.
func Reader(resolver func(id int64) (tg.InputUserClass, error), r io.Reader) styling.StyledTextOption {
	return styling.Custom(func(eb *entity.Builder) error {
		return Markdown(r, eb, Options{
			UserResolver: resolver,
		})
	})
}
