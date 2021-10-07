package unpack

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
)

func TestMessage(t *testing.T) {
	testMsg := &tg.Message{
		ID:      10,
		Message: "Golang is always going to do some approximation of the right thing.",
	}
	testErr := xerrors.New("женой накормили толпу")

	type args struct {
		u   tg.UpdatesClass
		err error
	}
	tests := []struct {
		name   string
		input  args
		output *tg.Message
		isErr  bool
	}{
		{"Good", args{
			u: &tg.Updates{
				Updates: []tg.UpdateClass{
					&tg.UpdateNewMessage{
						Message: testMsg,
					},
				},
			},
			err: nil,
		}, testMsg, false},
		{"GoodChannel", args{
			u: &tg.Updates{
				Updates: []tg.UpdateClass{
					&tg.UpdateNewChannelMessage{
						Message: testMsg,
					},
				},
			},
			err: nil,
		}, testMsg, false},
		{"Short", args{
			u: &tg.UpdateShort{
				Update: &tg.UpdateNewMessage{
					Message: testMsg,
				},
			},
			err: nil,
		}, testMsg, false},
		{"ShortSent", args{
			u: &tg.UpdateShortSentMessage{
				Out:       testMsg.Out,
				ID:        testMsg.ID,
				Date:      testMsg.Date,
				Media:     testMsg.Media,
				Entities:  testMsg.Entities,
				TTLPeriod: testMsg.TTLPeriod,
			},
			err: nil,
		}, testMsg, false},
		{"ShortChat", args{
			u: &tg.UpdateShortChatMessage{
				Out:         testMsg.Out,
				Mentioned:   testMsg.Mentioned,
				MediaUnread: testMsg.MediaUnread,
				Silent:      testMsg.Silent,
				ID:          testMsg.ID,
				Message:     testMsg.Message,
				Date:        testMsg.Date,
				FwdFrom:     testMsg.FwdFrom,
				ViaBotID:    testMsg.ViaBotID,
				ReplyTo:     testMsg.ReplyTo,
				Entities:    testMsg.Entities,
				TTLPeriod:   testMsg.TTLPeriod,
			},
			err: nil,
		}, testMsg, false},
		{"ExternalError", args{
			u:   nil,
			err: testErr,
		}, nil, true},
		{"BadUnpack", args{
			u: &tg.UpdateShort{
				Update: &tg.UpdateConfig{},
			},
			err: nil,
		}, nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := require.New(t)

			r, err := Message(test.input.u, test.input.err)
			if test.isErr {
				a.Error(err)
				return
			}

			a.NoError(err)
			a.Equal(testMsg.ID, r.ID)
		})
	}
}
