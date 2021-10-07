package bin_test

import (
	"testing"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/telegram/message/entity"
	"github.com/nnqq/td/telegram/message/styling"
	"github.com/nnqq/td/tg"
)

func BenchmarkDecodeSlice(b *testing.B) {
	builder := entity.Builder{}
	if err := styling.Perform(&builder,
		styling.Plain("plaintext"), styling.Plain("\n\n"),
		styling.Mention("@durov"), styling.Plain("\n\n"),
		styling.Hashtag("#hashtag"), styling.Plain("\n\n"),
		styling.BotCommand("/command"), styling.Plain("\n\n"),
		styling.URL("https://google.org"), styling.Plain("\n\n"),
		styling.Email("example@example.org"), styling.Plain("\n\n"),
		styling.Bold("bold"), styling.Plain("\n\n"),
		styling.Italic("italic"), styling.Plain("\n\n"),
		styling.Underline("underline"), styling.Plain("\n\n"),
		styling.Strike("strike"), styling.Plain("\n\n"),
		styling.Code("fmt.Println(`Hello, World!`)"), styling.Plain("\n\n"),
		styling.PreLang("fmt.Println(`Hello, World!`)", "Go"), styling.Plain("\n\n"),
		styling.TextURL("clickme", "https://google.com"), styling.Plain("\n\n"),
		styling.Phone("+71234567891"), styling.Plain("\n\n"),
		styling.Cashtag("$CASHTAG"), styling.Plain("\n\n"),
		styling.Blockquote("blockquote"), styling.Plain("\n\n"),
		styling.BankCard("5550111111111111"), styling.Plain("\n\n"),
	); err != nil {
		b.Fatal(err)
	}
	message, entities := builder.Complete()

	var buf bin.Buffer
	if err := buf.Encode(&tg.Message{
		PeerID:   &tg.PeerUser{},
		Message:  message,
		Entities: entities,
	}); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var msg tg.Message

		if err := msg.Decode(&bin.Buffer{Buf: buf.Buf}); err != nil {
			b.Fatal(err)
		}
	}
}
