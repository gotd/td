package peer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestConstraints(t *testing.T) {
	createPromise := func(peer tg.InputPeerClass) Promise {
		return func(ctx context.Context) (tg.InputPeerClass, error) {
			return peer, nil
		}
	}

	tests := []struct {
		name      string
		decorator func(Promise) Promise
		input     tg.InputPeerClass
		wantErr   bool
	}{
		{"Channel", OnlyChannel, &tg.InputPeerChannel{
			ChannelID:  10,
			AccessHash: 10,
		}, false},
		{"Channel", OnlyChannel, &tg.InputPeerUser{
			UserID:     10,
			AccessHash: 10,
		}, true},
		{"User", OnlyUser, &tg.InputPeerUser{
			UserID:     10,
			AccessHash: 10,
		}, false},
		{"User", OnlyUser, &tg.InputPeerChannel{
			ChannelID:  10,
			AccessHash: 10,
		}, true},
		{"Chat", OnlyChat, &tg.InputPeerChat{
			ChatID: 10,
		}, false},
		{"Chat", OnlyChat, &tg.InputPeerChannel{
			ChannelID:  10,
			AccessHash: 10,
		}, true},
	}
	for _, test := range tests {
		tname := "Good"
		if test.wantErr {
			tname = "Bad"
		}
		t.Run(tname, func(t *testing.T) {
			t.Run(test.name, func(t *testing.T) {
				a := require.New(t)
				promise := test.decorator(createPromise(test.input))
				_, err := promise(context.Background())
				if test.wantErr {
					a.Error(err)
				} else {
					a.NoError(err)
				}
			})
		})
	}
}
