package markdown

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
)

type testCase struct {
	input    string
	msg      string
	entities func(msg string) []tg.MessageEntityClass
	wantErr  bool
}

func getEntities(formats ...entity.Formatter) func(msg string) []tg.MessageEntityClass {
	return func(msg string) []tg.MessageEntityClass {
		length := entity.ComputeLength(msg)
		r := make([]tg.MessageEntityClass, len(formats))
		for i := range formats {
			r[i] = formats[i](0, length)
		}
		return r
	}
}

func runTests(tests []testCase) func(t *testing.T) {
	return func(t *testing.T) {
		for i, test := range tests {
			testName := test.msg
			if testName == "" {
				testName = fmt.Sprintf("Test%d", i+1)
			}
			t.Run(testName, func(t *testing.T) {
				t.Cleanup(func() {
					if t.Failed() {
						t.Logf("Input: %q", test.input)
					}
				})
				a := require.New(t)
				b := entity.Builder{}

				err := Markdown(strings.NewReader(test.input), &b, Options{})
				if test.wantErr {
					a.Error(err)
					return
				}
				a.NoError(err)

				msg, entities := b.Complete()
				a.Equal(test.msg, msg)
				if test.entities != nil {
					expect := test.entities(test.msg)
					a.Len(entities, len(expect))
					a.ElementsMatch(expect, entities)
				} else {
					a.Empty(entities)
				}
			})
		}
	}
}

func TestMarkdown(t *testing.T) {
	t.Run("Common", runTests([]testCase{
		{input: "_italic_", msg: "italic", entities: getEntities(entity.Italic())},
		{input: "*italic*", msg: "italic", entities: getEntities(entity.Italic())},
		{input: "**bold**", msg: "bold", entities: getEntities(entity.Bold())},
		{input: "__bold__", msg: "bold", entities: getEntities(entity.Bold())},
		{input: "~~strikethrough~~", msg: "strikethrough", entities: getEntities(entity.Strike())},
		{input: "`inline code`", msg: "inline code", entities: getEntities(entity.Code())},
		{input: "```\npre block\n```", msg: "pre block", entities: getEntities(entity.Pre(""))},
		{input: "```python\nprint(1)\n```", msg: "print(1)", entities: getEntities(entity.Pre("python"))},
		{
			input: "[inline URL](http://www.example.com/)", msg: "inline URL",
			entities: getEntities(entity.TextURL("http://www.example.com/")),
		},
		{
			input: "[inline mention of a user](tg://user?id=123456789)", msg: "inline mention of a user",
			entities: getEntities(entity.MentionName(&tg.InputUser{UserID: 123456789})),
		},
		{
			input: "![👍](tg://emoji?id=5368324170671202286)", msg: "👍",
			entities: getEntities(entity.CustomEmoji(5368324170671202286)),
		},
		{input: ">quote", msg: "quote", entities: getEntities(entity.Blockquote(false))},
		// Plain text.
		{input: "just text", msg: "just text"},
		{input: "", msg: ""},
	}))

	t.Run("Spoiler", runTests([]testCase{
		{input: "||spoiler||", msg: "spoiler", entities: getEntities(entity.Spoiler())},
		{
			input: "before ||hidden|| after", msg: "before hidden after",
			entities: func(msg string) []tg.MessageEntityClass {
				return []tg.MessageEntityClass{
					&tg.MessageEntitySpoiler{Offset: 7, Length: 6},
				}
			},
		},
		{
			input: "||spoiler with *bold*||", msg: "spoiler with bold",
			entities: func(msg string) []tg.MessageEntityClass {
				return []tg.MessageEntityClass{
					&tg.MessageEntitySpoiler{Offset: 0, Length: 17},
					&tg.MessageEntityItalic{Offset: 13, Length: 4},
				}
			},
		},
		// A single pipe is not a spoiler delimiter.
		{input: "a | b", msg: "a | b"},
		// Escaped pipes are literal.
		{input: `\|\|not spoiler\|\|`, msg: "||not spoiler||"},
	}))

	t.Run("Nested", runTests([]testCase{
		{
			input: "**bold _italic_**", msg: "bold italic",
			entities: func(msg string) []tg.MessageEntityClass {
				return []tg.MessageEntityClass{
					&tg.MessageEntityBold{Offset: 0, Length: 11},
					&tg.MessageEntityItalic{Offset: 5, Length: 6},
				}
			},
		},
		{
			input: "**bold** and _italic_", msg: "bold and italic",
			entities: func(msg string) []tg.MessageEntityClass {
				return []tg.MessageEntityClass{
					&tg.MessageEntityBold{Offset: 0, Length: 4},
					&tg.MessageEntityItalic{Offset: 9, Length: 6},
				}
			},
		},
	}))

	t.Run("Escape", runTests([]testCase{
		{input: `\*not italic\*`, msg: "*not italic*"},
		{input: `\_\_\_`, msg: "___"},
		{input: `2 \> 1`, msg: "2 > 1"},
		{input: `a\.b\.c`, msg: "a.b.c"},
	}))

	t.Run("Blockquote", runTests([]testCase{
		{
			input: ">line one\n>line two", msg: "line one\nline two",
			entities: getEntities(entity.Blockquote(false)),
		},
		{
			// Blank line terminates the blockquote (CommonMark).
			input: ">quoted\n\nplain", msg: "quoted\n\nplain",
			entities: func(msg string) []tg.MessageEntityClass {
				return []tg.MessageEntityClass{
					&tg.MessageEntityBlockquote{Offset: 0, Length: 6},
				}
			},
		},
	}))

	// The parser is lenient: unmatched markup is emitted as plain text.
	t.Run("Lenient", runTests([]testCase{
		{input: "*unclosed italic", msg: "*unclosed italic"},
		{input: "just a | pipe", msg: "just a | pipe"},
		{input: "closing ] only", msg: "closing ] only"},
		{input: `bad escape \a`, msg: `bad escape \a`},
	}))
}
