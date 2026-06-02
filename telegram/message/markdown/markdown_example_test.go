package markdown_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/markdown"
	"github.com/gotd/td/tg"
)

func sendMarkdown(ctx context.Context) error {
	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return err
	}

	// This example creates a styled message from Markdown
	// and sends it to your Saved Messages folder.
	return client.Run(ctx, func(ctx context.Context) error {
		_, err := message.NewSender(tg.NewClient(client)).
			Self().StyledText(ctx, markdown.String(nil, `**bold text**
_italic text_
~~strikethrough~~
**bold _italic bold_ bold**
[inline URL](http://www.example.com/)
[inline mention of a user](tg://user?id=123456789)
![👍](tg://emoji?id=5368324170671202286)
`+"`inline fixed-width code`"+`
`+"```python"+`
pre-formatted fixed-width code block written in the Python programming language
`+"```"+`
> Block quotation started
> The last line of the block quotation`))
		return err
	})
}

func ExampleString() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := sendMarkdown(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
