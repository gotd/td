package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestManager_findPeerClass(t *testing.T) {
	user := getTestSelf()
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
				p:     &tg.PeerUser{UserID: user.GetID()},
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
	testUser := getTestSelf()
	inputs := []struct {
		Name  string
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
				Peer: &tg.PeerUser{UserID: testUser.GetID()},
				Users: []tg.UserClass{
					&tg.User{ID: testUser.GetID(), AccessHash: 10, Username: username},
				},
			}).ExpectCall(&tg.ContactsResolveUsernameRequest{
				Username: username,
			}).ThenRPCErr(getTestError())

			ctx := context.Background()

			r, err := m.Resolve(ctx, tt.Input)
			a.NoError(err)
			a.IsType(&tg.InputPeerUser{}, r.InputPeer())
			a.Equal(testUser.GetID(), r.ID())

			_, err = m.Resolve(ctx, tt.Input)
			a.Error(err)
		})
	}
}

func TestManager_ResolvePhone(t *testing.T) {
	a := require.New(t)
	mock, m := testManager(t)
	// To count contacts hash.
	m.me.Store(&tg.User{
		ID: 1,
	})

	phone := "+79001234567"
	phone2 := "+79011234567"
	mock.ExpectCall(&tg.ContactsGetContactsRequest{
		Hash: 0,
	}).ThenRPCErr(getTestError())
	ctx := context.Background()

	_, err := m.Resolve(ctx, phone)
	a.Error(err)

	resp := &tg.ContactsContacts{
		Contacts: []tg.Contact{{
			UserID: 10,
			Mutual: false,
		}},
		SavedCount: 1,
		Users: []tg.UserClass{
			&tg.User{ID: 10, AccessHash: 10, Username: "rustmustdie", Phone: cleanupPhone(phone)},
		},
	}
	mock.ExpectCall(&tg.ContactsGetContactsRequest{
		Hash: 0,
	}).ThenResult(resp)

	r, err := m.Resolve(ctx, phone)
	a.NoError(err)
	a.IsType(&tg.InputPeerUser{}, r.InputPeer())
	a.Equal(int64(10), r.ID())

	mock.ExpectCall(&tg.ContactsGetContactsRequest{
		Hash: contactsHash(1, resp),
	}).ThenResult(&tg.ContactsContactsNotModified{})

	_, err = m.Resolve(ctx, phone2)
	a.Error(err)
}

func TestManager_ResolveBusinessChat(t *testing.T) {
	const slug = "slug"
	testUser := getTestUser()

	apiReq := &tg.AccountResolveBusinessChatLinkRequest{Slug: slug}
	apiResp := &tg.AccountResolvedBusinessChatLinks{
		Peer: &tg.PeerUser{UserID: testUser.GetID()},
		Users: []tg.UserClass{
			&tg.User{ID: testUser.GetID(), AccessHash: 10, Username: testUser.Username},
		},
		Message: slug,
	}

	inputs := []struct {
		Name         string
		Input        string
		wantParseErr bool
		wantApiErr   bool
	}{
		{"Business chat link", "https://t.me/m/" + slug, false, false},
		{"Api error", "https://t.me/m/" + slug, false, true},
		{"Not business chat link", "https://t.me/not_business_link/", true, false},
		{"No slug", "https://t.me/m/", true, false},
	}

	for _, tt := range inputs {
		t.Run(tt.Name, func(t *testing.T) {
			a := require.New(t)
			mock, m := testManager(t)

			if !tt.wantParseErr {
				rb := mock.ExpectCall(apiReq)
				if tt.wantApiErr {
					rb.ThenRPCErr(getTestError())
				} else {
					rb.ThenResult(apiResp)
				}
			}

			r, msg, err := m.ResolveBusinessChat(context.Background(), tt.Input)
			if tt.wantParseErr || tt.wantApiErr {
				a.Error(err)
				return
			}

			a.NoError(err)
			a.Equal(slug, msg.Msg)

			a.IsType(&tg.InputPeerUser{}, r.InputPeer())
			a.Equal(testUser.GetID(), r.ID())
		})
	}
}
