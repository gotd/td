package rich

import (
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Custom inline AST node kinds for the rich-only Markdown syntax that goldmark
// does not parse natively: ==marked==, ||spoiler|| and $math$.
var (
	kindMarked  = gast.NewNodeKind("RichMarked")
	kindSpoiler = gast.NewNodeKind("RichSpoiler")
	kindMath    = gast.NewNodeKind("RichMath")
)

type markedNode struct{ gast.BaseInline }

func (n *markedNode) Kind() gast.NodeKind        { return kindMarked }
func (n *markedNode) Dump(src []byte, level int) { gast.DumpHelper(n, src, level, nil, nil) }

type spoilerNode struct{ gast.BaseInline }

func (n *spoilerNode) Kind() gast.NodeKind        { return kindSpoiler }
func (n *spoilerNode) Dump(src []byte, level int) { gast.DumpHelper(n, src, level, nil, nil) }

type mathNode struct {
	gast.BaseInline
	Source  string
	Display bool
}

func (n *mathNode) Kind() gast.NodeKind        { return kindMath }
func (n *mathNode) Dump(src []byte, level int) { gast.DumpHelper(n, src, level, nil, nil) }

// pairDelimiterProcessor matches a paired two-character delimiter (== or ||).
type pairDelimiterProcessor struct{ char byte }

func (p *pairDelimiterProcessor) IsDelimiter(b byte) bool { return b == p.char }

func (p *pairDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *pairDelimiterProcessor) OnMatch(int) gast.Node {
	if p.char == '=' {
		return &markedNode{}
	}
	return &spoilerNode{}
}

// pairParser parses a two-character paired delimiter such as ==text== (marked)
// or ||text|| (spoiler).
type pairParser struct {
	char byte
	proc *pairDelimiterProcessor
}

func newPairParser(char byte) *pairParser {
	return &pairParser{char: char, proc: &pairDelimiterProcessor{char: char}}
}

func (s *pairParser) Trigger() []byte { return []byte{s.char} }

func (s *pairParser) Parse(_ gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 2, s.proc)
	// Only a run of exactly two delimiter characters opens/closes the span.
	if node == nil || node.OriginalLength != 2 {
		return nil
	}
	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

func (s *pairParser) CloseBlock(gast.Node, parser.Context) {}

// mathParser parses inline $...$ and display $$...$$ mathematical expressions
// whose content is literal LaTeX.
type mathParser struct{}

func (m *mathParser) Trigger() []byte { return []byte{'$'} }

func (m *mathParser) Parse(_ gast.Node, block text.Reader, _ parser.Context) gast.Node {
	line, _ := block.PeekLine()
	if len(line) < 2 || line[0] != '$' {
		return nil
	}
	display := line[1] == '$'
	open := 1
	if display {
		open = 2
	}

	rest := line[open:]
	closeIdx := -1
	for i := 0; i+open <= len(rest); i++ {
		if rest[i] != '$' {
			continue
		}
		if open == 1 || (i+1 < len(rest) && rest[i+1] == '$') {
			closeIdx = i
			break
		}
	}
	// Require non-empty content and a closing delimiter on the same line.
	if closeIdx <= 0 {
		return nil
	}

	source := string(rest[:closeIdx])
	block.Advance(open + closeIdx + open)
	return &mathNode{Source: source, Display: display}
}
