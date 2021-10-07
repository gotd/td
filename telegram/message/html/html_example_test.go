package html_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/nnqq/td/telegram"
	"github.com/nnqq/td/telegram/message"
	"github.com/nnqq/td/telegram/message/html"
	"github.com/nnqq/td/tg"
)

func sendHTML(ctx context.Context) error {
	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return err
	}

	// This example creates a styled message from BotAPI examples
	// and sends to your Saved Messages folder.
	// See https://core.telegram.org/bots/api#html-style.
	return client.Run(ctx, func(ctx context.Context) error {
		_, err := message.NewSender(tg.NewClient(client)).
			Self().StyledText(ctx, html.String(nil, `<b>bold</b>, <strong>bold</strong>
<i>italic</i>, <em>italic</em>
<u>underline</u>, <ins>underline</ins>
<s>strikethrough</s>, <strike>strikethrough</strike>, <del>strikethrough</del>
<b>bold <i>italic bold <s>italic bold strikethrough</s> <u>underline italic bold</u></i> bold</b>
<a href="http://www.example.com/">inline URL</a>
<a href="tg://user?id=123456789">inline mention of a user</a>
<code>inline fixed-width code</code>
<pre>pre-formatted fixed-width code block</pre>
<pre><code class="language-python">pre-formatted fixed-width code block written in the Python programming language</code></pre>`))
		return err
	})
}

func ExampleString() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := sendHTML(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
