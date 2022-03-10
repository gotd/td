package members

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestChatMembers_Count(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	now := time.Now()
	date := int(now.Unix())

	rawCh := getTestChat()
	rawCh.Date = date
	ch := m.Chat(rawCh)
	members := Chat(ch)

	mock.ExpectCall(&tg.MessagesGetFullChatRequest{
		ChatID: ch.ID(),
	}).ThenRPCErr(getTestError())
	_, err := members.Count(ctx)
	a.Error(err)

	mock.ExpectCall(&tg.MessagesGetFullChatRequest{
		ChatID: ch.ID(),
	}).ThenResult(&tg.MessagesChatFull{
		FullChat: getTestChatFull(&tg.ChatParticipants{
			ChatID: 10,
			Participants: []tg.ChatParticipantClass{
				&tg.ChatParticipant{
					UserID:    10,
					InviterID: 11,
					Date:      date,
				},
				&tg.ChatParticipantCreator{
					UserID: 10,
				},
				&tg.ChatParticipantAdmin{
					UserID:    10,
					InviterID: 11,
					Date:      date,
				},
			},
			Version: 1,
		}),
	})
	count, err := members.Count(ctx)
	a.Equal(3, count)
	a.NoError(err)
}

func TestChatMembers_ForEach(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	now := time.Now()
	date := int(now.Unix())

	rawCh := getTestChat()
	rawCh.Date = date
	ch := m.Chat(rawCh)

	mock.ExpectCall(&tg.MessagesGetFullChatRequest{
		ChatID: ch.ID(),
	}).ThenResult(&tg.MessagesChatFull{
		FullChat: getTestChatFull(&tg.ChatParticipants{
			ChatID: 10,
			Participants: []tg.ChatParticipantClass{
				&tg.ChatParticipant{
					UserID:    10,
					InviterID: 11,
					Date:      date,
				},
				&tg.ChatParticipantCreator{
					UserID: 10,
				},
				&tg.ChatParticipantAdmin{
					UserID:    10,
					InviterID: 11,
					Date:      date,
				},
			},
			Version: 1,
		}),
		Users: []tg.UserClass{
			&tg.User{
				ID:         10,
				AccessHash: 10,
			},
			&tg.User{
				ID:         11,
				AccessHash: 10,
			},
		},
	})
	members := Chat(ch)

	count, err := members.Count(ctx)
	a.Equal(3, count)
	a.NoError(err)

	expected := []struct {
		Status      Status
		JoinDate    time.Time
		JoinDateSet bool
		InviterID   int64
	}{
		{Status: Plain, JoinDate: now, JoinDateSet: true, InviterID: 11},
		{Status: Creator, JoinDate: now, JoinDateSet: true},
		{Status: Admin, JoinDate: now, JoinDateSet: true, InviterID: 11},
	}

	i := 0
	a.NoError(members.ForEach(ctx, func(m Member) error {
		p := m.(ChatMember)
		e := expected[i]

		a.Equal(e.Status, p.Status(), i)
		a.Equal(int64(10), p.User().ID())
		if join, ok := p.JoinDate(); e.JoinDateSet {
			a.True(ok, i)
			a.Equal(e.JoinDate.Unix(), join.Unix(), i)
		} else {
			a.False(ok, i)
		}

		if inviter, ok := p.InvitedBy(); e.InviterID != 0 {
			a.True(ok, i)
			a.Equal(e.InviterID, inviter.ID())
		} else {
			a.False(ok, i)
		}

		i++
		return nil
	}))
}

func TestChatMembers_Kick(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	u := m.User(getTestUser())
	ch := m.Chat(getTestChat())
	members := Chat(ch)

	mock.ExpectCall(&tg.MessagesDeleteChatUserRequest{
		RevokeHistory: true,
		ChatID:        ch.ID(),
		UserID:        u.InputUser(),
	}).ThenRPCErr(getTestError())
	a.Error(members.Kick(ctx, u.InputUser(), true))

	mock.ExpectCall(&tg.MessagesDeleteChatUserRequest{
		RevokeHistory: true,
		ChatID:        ch.ID(),
		UserID:        u.InputUser(),
	}).ThenResult(&tg.Updates{})
	a.NoError(members.Kick(ctx, u.InputUser(), true))
}

func TestChatMembers_EditAdmin(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	u := m.User(getTestUser())
	ch := m.Chat(getTestChat())
	members := Chat(ch)

	mock.ExpectCall(&tg.MessagesEditChatAdminRequest{
		IsAdmin: true,
		ChatID:  ch.ID(),
		UserID:  u.InputUser(),
	}).ThenRPCErr(getTestError())
	a.Error(members.EditAdmin(ctx, u.InputUser(), true))

	mock.ExpectCall(&tg.MessagesEditChatAdminRequest{
		IsAdmin: true,
		ChatID:  ch.ID(),
		UserID:  u.InputUser(),
	}).ThenTrue()
	a.NoError(members.EditAdmin(ctx, u.InputUser(), true))

	mock.ExpectCall(&tg.MessagesEditChatAdminRequest{
		IsAdmin: false,
		ChatID:  ch.ID(),
		UserID:  u.InputUser(),
	}).ThenTrue()
	a.NoError(members.EditAdmin(ctx, u.InputUser(), false))
}
