package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func TestManager_ImportInvite(t *testing.T) {
	ctx := context.Background()
	hash := "aboba"
	expectCheck := func(m *tgmock.Mock) *tgmock.RequestBuilder {
		return m.ExpectCall(&tg.MessagesCheckChatInviteRequest{
			Hash: hash,
		})
	}
	expectResult := func(m *tgmock.Mock, class tg.ChatInviteClass) *tgmock.Mock {
		return expectCheck(m).ThenResult(class)
	}

	t.Run("CheckError", func(t *testing.T) {
		a := require.New(t)
		mock, m := testManager(t)

		expectCheck(mock).ThenRPCErr(getTestError())
		_, err := m.ImportInvite(ctx, hash)
		a.Error(err)
	})
	t.Run("ChatInviteAlready", func(t *testing.T) {
		a := require.New(t)
		mock, m := testManager(t)

		testChat := getTestChannel()
		expectResult(mock, &tg.ChatInviteAlready{
			Chat: testChat,
		})
		r, err := m.ImportInvite(ctx, hash)
		a.NoError(err)
		a.Equal(testChat.ID, r.ID())
	})

	testImport := func(testChat *tg.Chat, invite tg.ChatInviteClass) func(t *testing.T) {
		return func(t *testing.T) {
			a := require.New(t)
			mock, m := testManager(t)

			expectResult(mock, invite).ExpectCall(&tg.MessagesImportChatInviteRequest{
				Hash: hash,
			}).ThenRPCErr(getTestError())
			_, err := m.ImportInvite(ctx, hash)
			a.Error(err)

			expectResult(mock, invite).ExpectCall(&tg.MessagesImportChatInviteRequest{
				Hash: hash,
			}).ThenResult(&tg.Updates{
				Chats: []tg.ChatClass{testChat},
			})
			r, err := m.ImportInvite(ctx, hash)
			a.NoError(err)
			a.Equal(testChat.ID, r.ID())
		}
	}

	testChat := getTestChat()
	t.Run("ChatInvite", testImport(testChat, &tg.ChatInvite{
		Channel:           false,
		Broadcast:         false,
		Public:            false,
		Megagroup:         false,
		RequestNeeded:     false,
		Title:             testChat.Title,
		About:             "",
		Photo:             &tg.PhotoEmpty{},
		ParticipantsCount: testChat.ParticipantsCount,
	}))
	t.Run("ChatInvitePeek", testImport(testChat, &tg.ChatInvitePeek{
		Chat: testChat,
	}))
}
