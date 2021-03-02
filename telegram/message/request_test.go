package message

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPure(t *testing.T) {
	s := Sender{}
	dialog := s.Self()

	b1 := dialog.Reply(1).Reply(2).Reply(1)
	b2 := dialog.Reply(2)
	b3 := dialog

	require.Equal(t, 1, b1.replyToMsgID)
	require.Equal(t, 2, b2.replyToMsgID)
	require.Equal(t, 0, b3.replyToMsgID)
}
