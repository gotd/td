package message

import (
	"context"
	"fmt"
	"testing"

	"github.com/gotd/td/tg"
)

func TestSender_JoinLink(t *testing.T) {
	formats := []struct {
		fmt     string
		wantErr bool
	}{
		{`t.me/joinchat/%s`, false},
		{`t.me/joinchat/%s/`, false},
		{`https://t.me/joinchat/%s`, false},
		{`https://t.me/joinchat/%s/`, false},
		{`tg:join?invite=%s`, false},
		{`tg://join?invite=%s`, false},
	}
	inputs := []struct {
		value   string
		wantErr bool
	}{
		{"AAAAAAAAAAAAAAAAAA", false},
		{"", true},
	}

	for _, format := range formats {
		for _, input := range inputs {
			link := fmt.Sprintf(format.fmt, input.value)
			t.Run(link, func(t *testing.T) {
				ctx := context.Background()
				sender, mock := testSender(t)

				wantErr := format.wantErr || input.wantErr
				if !wantErr {
					mock.ExpectCall(&tg.MessagesImportChatInviteRequest{
						Hash: input.value,
					}).ThenResult(&tg.Updates{})
				}

				_, err := sender.JoinLink(ctx, link)
				if wantErr {
					mock.Error(err)
				} else {
					mock.NoError(err)
				}
			})
		}
	}
}
