// Binary deeplink parses Telegram deeplinks (t.me / tg: links) using the
// telegram/deeplink helper.
//
// Unlike the other examples it does not connect to Telegram: deeplink parsing
// is fully offline. It is handy for routing incoming links (resolve a user,
// join a chat by invite, open a business chat, ...).
//
// Usage:
//
//	go run ./deeplink https://t.me/gotd
//	go run ./deeplink "tg:join?invite=AAAAAA"
package main

import (
	"fmt"
	"os"

	"github.com/gotd/td/telegram/deeplink"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: deeplink <link>")
		os.Exit(2)
	}
	link := os.Args[1]

	// IsDeeplinkLike is a cheap pre-check, useful to decide whether a string
	// is worth parsing at all.
	if !deeplink.IsDeeplinkLike(link) {
		fmt.Printf("%q does not look like a deeplink\n", link)
		os.Exit(1)
	}

	d, err := deeplink.Parse(link)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("type: %s\n", d.Type)
	for key := range d.Args {
		fmt.Printf("  %s = %s\n", key, d.Args.Get(key))
	}

	// Switch on the deeplink type to decide what to do next.
	switch d.Type {
	case deeplink.Resolve:
		fmt.Printf("=> resolve username %q (e.g. contacts.resolveUsername)\n", d.Args.Get("domain"))
	case deeplink.Join:
		fmt.Printf("=> join chat by invite %q (e.g. messages.importChatInvite)\n", d.Args.Get("invite"))
	case deeplink.BusinessChat:
		fmt.Printf("=> open business chat %q\n", d.Args.Get("slug"))
	}
}
