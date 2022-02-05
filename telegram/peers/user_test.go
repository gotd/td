package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestUserGetters(t *testing.T) {
	a := require.New(t)
	u := User{
		raw: &tg.User{
			Self:           true,
			Contact:        true,
			MutualContact:  true,
			Deleted:        true,
			Bot:            true,
			BotChatHistory: true,
			BotNochats:     true,
			Verified:       true,
			Restricted:     true,
			Min:            true,
			BotInlineGeo:   true,
			Support:        true,
			Scam:           true,
			ApplyMinPhoto:  true,
			Fake:           true,
			ID:             10,
			AccessHash:     10,
			FirstName:      "FirstName",
			LastName:       "LastName",
			Username:       "Username",
			Phone:          "+79001234567",
			Photo: &tg.UserProfilePhoto{
				HasVideo:      true,
				PhotoID:       10,
				StrippedThumb: []byte("abc"),
				DCID:          1,
			},
			Status:         &tg.UserStatusLastMonth{},
			BotInfoVersion: 1,
			RestrictionReason: []tg.RestrictionReason{{
				Platform: "ios",
				Reason:   "ban",
				Text:     "ban",
			}},
			BotInlinePlaceholder: "placeholder",
			LangCode:             "ru",
		},
	}
	u.raw.SetFlags()
	a.Equal(u.raw, u.Raw())
	a.True(u.TDLibPeerID().IsUser())

	a.Equal(u.raw.GetSelf(), u.Self())
	a.Equal(u.raw.GetContact(), u.Contact())
	a.Equal(u.raw.GetMutualContact(), u.MutualContact())
	a.Equal(u.raw.GetDeleted(), u.Deleted())
	a.Equal(u.raw.GetVerified(), u.Verified())
	a.Equal(u.raw.GetSupport(), u.Support())
	a.Equal(u.raw.GetScam(), u.Scam())
	a.Equal(u.raw.GetFake(), u.Fake())
	a.Equal(u.raw.GetID(), u.ID())
	{
		reasons, ok := u.Restricted()
		a.Equal(u.raw.GetRestricted(), ok)
		a.Equal(u.raw.RestrictionReason, reasons)
	}
	{
		v, ok := u.raw.GetFirstName()
		v2, ok2 := u.FirstName()
		a.Equal(ok, ok2)
		a.Equal(v, v2)
	}
	{
		v, ok := u.raw.GetLastName()
		v2, ok2 := u.LastName()
		a.Equal(ok, ok2)
		a.Equal(v, v2)
	}
	{
		v, ok := u.raw.GetUsername()
		v2, ok2 := u.Username()
		a.Equal(ok, ok2)
		a.Equal(v, v2)
	}
	{
		v, ok := u.raw.GetPhone()
		v2, ok2 := u.Phone()
		a.Equal(ok, ok2)
		a.Equal(v, v2)
	}
	{
		v, ok := u.raw.GetStatus()
		v2, ok2 := u.Status()
		a.Equal(ok, ok2)
		a.Equal(v, v2)
	}
	{
		v, ok := u.raw.GetLangCode()
		v2, ok2 := u.LangCode()
		a.Equal(ok, ok2)
		a.Equal(v, v2)
	}

	b, ok := u.ToBot()
	a.True(ok)
	a.Equal(b.raw.GetBotChatHistory(), b.ChatHistory())
	a.Equal(!b.raw.GetBotNochats(), b.CanBeAdded())
	a.Equal(b.raw.GetBotInlineGeo(), b.InlineGeo())
	_, ok = b.raw.GetBotInlinePlaceholder()
	a.Equal(ok, b.SupportsInline())
}

func TestUser_InputPeer(t *testing.T) {
	require.Equal(t, &tg.InputPeerSelf{}, User{raw: &tg.User{Self: true}}.InputPeer())
	require.Equal(t, &tg.InputPeerUser{
		UserID:     10,
		AccessHash: 10,
	}, User{raw: &tg.User{
		ID:         10,
		AccessHash: 10,
	}}.InputPeer())
}

func TestUser_InputUser(t *testing.T) {
	require.Equal(t, &tg.InputUserSelf{}, User{raw: &tg.User{Self: true}}.InputUser())
	require.Equal(t, &tg.InputUser{
		UserID:     10,
		AccessHash: 10,
	}, User{raw: &tg.User{
		ID:         10,
		AccessHash: 10,
	}}.InputUser())
}

func TestUser_VisibleName(t *testing.T) {
	require.Equal(t, "FirstName", User{raw: &tg.User{FirstName: "FirstName"}}.VisibleName())
	require.Equal(t, "FirstName LastName", User{raw: &tg.User{
		FirstName: "FirstName",
		LastName:  "LastName",
	}}.VisibleName())
}

func TestUser_ReportSpam(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	u := m.User(getTestUser())

	mock.ExpectCall(&tg.MessagesReportSpamRequest{Peer: u.InputPeer()}).
		ThenRPCErr(getTestError())
	a.Error(u.ReportSpam(ctx))

	mock.ExpectCall(&tg.MessagesReportSpamRequest{Peer: u.InputPeer()}).
		ThenTrue()
	a.NoError(u.ReportSpam(ctx))
}

func TestUser_Block(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	u := m.User(getTestUser())

	mock.ExpectCall(&tg.ContactsBlockRequest{ID: u.InputPeer()}).
		ThenRPCErr(getTestError())
	a.Error(u.Block(ctx))

	mock.ExpectCall(&tg.ContactsBlockRequest{ID: u.InputPeer()}).
		ThenTrue()
	a.NoError(u.Block(ctx))
}

func TestUser_Unblock(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	u := m.User(getTestUser())

	mock.ExpectCall(&tg.ContactsUnblockRequest{ID: u.InputPeer()}).
		ThenRPCErr(getTestError())
	a.Error(u.Unblock(ctx))

	mock.ExpectCall(&tg.ContactsUnblockRequest{ID: u.InputPeer()}).
		ThenTrue()
	a.NoError(u.Unblock(ctx))
}
