package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func TestManager_findPeerClass(t *testing.T) {
	user := &tg.User{
		ID: 10,
	}
	chat := &tg.Chat{
		ID: 10,
	}
	channel := &tg.Channel{
		ID: 10,
	}

	m := &Manager{}
	type args struct {
		p     tg.PeerClass
		users []tg.UserClass
		chats []tg.ChatClass
	}
	tests := []struct {
		name   string
		args   args
		want   Peer
		wantOk bool
	}{
		{
			name: "User",
			args: args{
				p:     &tg.PeerUser{UserID: 10},
				users: []tg.UserClass{user},
			},
			want: User{
				raw: user,
				m:   m,
			},
			wantOk: true,
		},
		{
			name: "Chat",
			args: args{
				p:     &tg.PeerChat{ChatID: 10},
				chats: []tg.ChatClass{chat},
			},
			want: Chat{
				raw: chat,
				m:   m,
			},
			wantOk: true,
		},
		{
			name: "Channel",
			args: args{
				p:     &tg.PeerChannel{ChannelID: 10},
				chats: []tg.ChatClass{channel},
			},
			want: Channel{
				raw: channel,
				m:   m,
			},
			wantOk: true,
		},
		{name: "NilPeer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)
			r, ok := m.findPeerClass(tt.args.p, tt.args.users, tt.args.chats)
			if tt.wantOk {
				a.Equal(tt.want, r)
				a.True(ok)
			} else {
				a.False(ok)
			}
		})
	}
}

func TestManager_Resolve(t *testing.T) {
	inputs := []struct{
		Name string
		Input string
	}{
		{"Domain", "@gotduser"},
		{"Deeplink", "https://t.me/gotduser"},
	}
	for _, tt := range inputs {
		t.Run(tt.Name, func(t *testing.T) {
			a := require.New(t)
			mock, m := testManager(t)

			username := "gotduser"
			mock.ExpectCall(&tg.ContactsResolveUsernameRequest{
				Username: username,
			}).ThenResult(&tg.ContactsResolvedPeer{
				Peer: &tg.PeerUser{UserID: 10},
				Users: []tg.UserClass{
					&tg.User{ID: 10, AccessHash: 10, Username: username},
				},
			}).ExpectCall(&tg.ContactsResolveUsernameRequest{
				Username: username,
			}).ThenRPCErr(&tgerr.Error{
				Code:    1337,
				Message: "TEST_ERROR",
				Type:    "TEST_ERROR",
			})

			ctx := context.Background()

			r, err := m.Resolve(ctx, tt.Input)
			a.NoError(err)
			a.IsType(&tg.InputPeerUser{}, r.InputPeer())
			a.Equal(int64(10), r.ID())

			_, err = m.Resolve(ctx, tt.Input)
			a.Error(err)
		})
	}
}

func TestManager_ResolvePhone(t *testing.T) {
	a := require.New(t)
	mock, m := testManager(t)

	phone := "+79001234567"
	mock.ExpectCall(&tg.ContactsGetContactsRequest{
		Hash: 0,
	}).ThenRPCErr(&tgerr.Error{
		Code:    1337,
		Message: "TEST_ERROR",
		Type:    "TEST_ERROR",
	})
	ctx := context.Background()

	_, err := m.Resolve(ctx, phone)
	a.Error(err)

	mock.ExpectCall(&tg.ContactsGetContactsRequest{
		Hash: 0,
	}).ThenResult(&tg.ContactsContacts{
		Contacts: []tg.Contact{{
			UserID: 10,
			Mutual: false,
		}},
		SavedCount: 1,
		Users: []tg.UserClass{
			&tg.User{ID: 10, AccessHash: 10, Username: "rustmustdie", Phone: cleanupPhone(phone)},
		},
	})

	r, err := m.Resolve(ctx, phone)
	a.NoError(err)
	a.IsType(&tg.InputPeerUser{}, r.InputPeer())
	a.Equal(int64(10), r.ID())
}
