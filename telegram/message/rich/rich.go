// Package rich builds Telegram rich messages.
//
// Rich messages (layer 227, Bot API 10.1) carry highly structured content:
// headings, paragraphs, lists, tables, block quotes, media, anchors, footnotes,
// mathematical expressions and more. Unlike ordinary messages — which are a
// flat string plus a list of [tg.MessageEntityClass] ranges — a rich message is
// a tree of [tg.PageBlockClass] blocks whose inline text is a tree of
// [tg.RichTextClass] nodes (the same model Telegram uses for Instant View
// pages).
//
// This package provides three things:
//
//   - Builder helpers for every RichText and PageBlock constructor, e.g.
//     [Bold], [Subscript], [Math], [Paragraph], [Heading1], [Table].
//   - [Message], which assembles a [tg.InputRichMessage] from blocks.
//   - [HTML] and [Markdown], which wrap an HTML or Markdown source into a
//     [tg.InputRichMessageHTML] / [tg.InputRichMessageMarkdown] for server-side
//     parsing, and [ParseHTML] / [ParseMarkdown], which parse a useful subset
//     locally into blocks.
//
// The result is a [tg.InputRichMessageClass] that can be sent or edited through
// the message sender (see (*message.Builder).RichMessage).
package rich

import "github.com/gotd/td/tg"

// join collapses a list of rich texts into a single RichText node, wrapping
// multiple nodes in a textConcat and returning textEmpty for none.
func join(texts []tg.RichTextClass) tg.RichTextClass {
	switch len(texts) {
	case 0:
		return &tg.TextEmpty{}
	case 1:
		return texts[0]
	default:
		return &tg.TextConcat{Texts: texts}
	}
}
