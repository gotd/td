package peer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func createPromise(peer tg.InputPeerClass) Promise {
	return func(ctx context.Context) (tg.InputPeerClass, error) {
		return peer, nil
	}
}

func TestConstraintsCombine(t *testing.T) {
	decorate := func(promise Promise, decorators ...PromiseDecorator) Promise {
		for _, decorator := range decorators {
			promise = decorator(promise)
		}
		return promise
	}

	tests := []struct {
		name       string
		decorators []PromiseDecorator
		input      tg.InputPeerClass
		wantErr    bool
	}{
		{
			"UserOrChannel",
			[]PromiseDecorator{OnlyUser, OnlyChannel},
			&tg.InputPeerChannel{
				ChannelID:  10,
				AccessHash: 10,
			},
			false,
		},
		{
			"UserOrChannel",
			[]PromiseDecorator{OnlyUser, OnlyChannel},
			&tg.InputPeerChat{
				ChatID: 10,
			},
			true,
		},
		{
			"ChannelOrUser",
			[]PromiseDecorator{OnlyChannel, OnlyUser},
			&tg.InputPeerUser{
				UserID:     10,
				AccessHash: 10,
			},
			false,
		},
		{
			"ChannelOrUser",
			[]PromiseDecorator{OnlyChannel, OnlyUser},
			&tg.InputPeerChat{
				ChatID: 10,
			},
			true,
		},
	}
	for _, test := range tests {
		tname := "Good"
		if test.wantErr {
			tname = "Bad"
		}
		t.Run(test.name, func(t *testing.T) {
			t.Run(tname, func(t *testing.T) {
				a := require.New(t)
				promise := decorate(createPromise(test.input), test.decorators...)
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

func TestConstraints(t *testing.T) {
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
		{"User", OnlyUser, &tg.InputPeerSelf{}, false},
		{"User", OnlyUser, &tg.InputPeerChannel{
			ChannelID:  10,
			AccessHash: 10,
		}, true},
		{"UserID", OnlyUserID, &tg.InputPeerUser{
			UserID:     10,
			AccessHash: 10,
		}, false},
		{"UserID", OnlyUserID, &tg.InputPeerChannel{
			ChannelID:  10,
			AccessHash: 10,
		}, true},
		{"UserID", OnlyUserID, &tg.InputPeerSelf{}, true},
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
		t.Run(test.name, func(t *testing.T) {
			t.Run(tname, func(t *testing.T) {
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
