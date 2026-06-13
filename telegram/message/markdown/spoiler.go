package markdown

import (
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// KindSpoiler is a NodeKind of a Spoiler node.
//
// Spoiler is a Telegram MarkdownV2 construct (||text||) that has no CommonMark
// equivalent; it maps onto the message spoiler entity.
var KindSpoiler = gast.NewNodeKind("Spoiler")

// spoiler is an inline node representing ||spoiler|| text.
type spoiler struct {
	gast.BaseInline
}

func (n *spoiler) Kind() gast.NodeKind { return KindSpoiler }

func (n *spoiler) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

type spoilerDelimiterProcessor struct{}

func (p *spoilerDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '|'
}

func (p *spoilerDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *spoilerDelimiterProcessor) OnMatch(consumes int) gast.Node {
	return &spoiler{}
}

var defaultSpoilerDelimiterProcessor = &spoilerDelimiterProcessor{}

type spoilerParser struct{}

// NewSpoilerParser returns a new InlineParser that parses ||spoiler||
// expressions.
func NewSpoilerParser() parser.InlineParser {
	return &spoilerParser{}
}

func (s *spoilerParser) Trigger() []byte {
	return []byte{'|'}
}

func (s *spoilerParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	// Telegram spoilers use exactly "||"; a single '|' or a longer run is not a
	// delimiter.
	node := parser.ScanDelimiter(line, before, 2, defaultSpoilerDelimiterProcessor)
	if node == nil || node.OriginalLength != 2 || before == '|' {
		return nil
	}

	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

func (s *spoilerParser) CloseBlock(parent gast.Node, pc parser.Context) {
	// nothing to do
}
