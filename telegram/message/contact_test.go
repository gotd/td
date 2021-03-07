package message

import (
	"context"
	"testing"
	"unicode/utf8"

	"github.com/gotd/td/tg"
)

func TestContact(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	contact := tg.InputMediaContact{
		FirstName:   "Михал Палыч",
		LastName:    "Терентьев",
		PhoneNumber: "22 505",
	}

	expectSendMediaAndText(&contact, mock, "че с деньгами?", &tg.MessageEntityBold{
		Length: utf8.RuneCountInString("че с деньгами?"),
	})
	_, err := sender.Self().Media(ctx, Contact(contact, Bold("че с деньгами?")))
	mock.NoError(err)
}
