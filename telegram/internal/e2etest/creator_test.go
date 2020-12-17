package e2etest

import "testing"

func Test_parseTokenMessage(t *testing.T) {
	token, err := parseTokenMessage(`Done! Congratulations on your new bot. You will find it at t.me/example_bot. You can now add a description, about section and profile picture for your bot, see /help for a list of commands. By the way, when you've finished creating your cool bot, ping our Bot Support if you want a better username for it. Just make sure the bot is fully operational before you do this.

Use this token to access the HTTP API:
1111111111:AAAAAA_BBBBBBBBBBBBBBBBBBBBBBBBBBBB
Keep your token secure and store it safely, it can be used by anyone to control your bot.

For a description of the Bot API, see this page: https://core.telegram.org/bots/api`)
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	if token != "1111111111:AAAAAA_BBBBBBBBBBBBBBBBBBBBBBBBBBBB" {
		t.Fatalf("unexpected token %s", token)
	}
}
