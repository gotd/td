package peers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func testExportLinkOptions() ExportLinkOptions {
	return ExportLinkOptions{
		RequestNeeded: true,
		ExpireDate:    time.Now(),
		UsageLimit:    1,
		Title:         "Title",
	}
}

func testChatInviteExported() tg.ChatInviteExported {
	opts := testExportLinkOptions()
	r := tg.ChatInviteExported{
		AdminID:       getTestUser().ID,
		RequestNeeded: opts.RequestNeeded,
		ExpireDate:    int(opts.ExpireDate.Unix()),
		UsageLimit:    opts.UsageLimit,
		Title:         opts.Title,
	}
	r.SetFlags()
	return r
}

func TestInviteLinks_newLink(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	opts := testExportLinkOptions()
	ch := m.Channel(getTestChannel())
	links := ch.InviteLinks()

	mock.ExpectCall(&tg.MessagesExportChatInviteRequest{
		LegacyRevokePermanent: false,
		RequestNeeded:         opts.RequestNeeded,
		Peer:                  ch.InputPeer(),
		ExpireDate:            int(opts.ExpireDate.Unix()),
		UsageLimit:            opts.UsageLimit,
		Title:                 opts.Title,
	}).ThenRPCErr(getTestError())
	_, err := links.AddNew(ctx, opts)
	a.Error(err)

	result := testChatInviteExported()
	mock.ExpectCall(&tg.MessagesExportChatInviteRequest{
		LegacyRevokePermanent: true,
		RequestNeeded:         opts.RequestNeeded,
		Peer:                  ch.InputPeer(),
		ExpireDate:            int(opts.ExpireDate.Unix()),
		UsageLimit:            opts.UsageLimit,
		Title:                 opts.Title,
	}).ThenResult(&result)
	got, err := links.ExportNew(ctx, opts)
	a.NoError(err)
	a.Equal(result, got.Raw())
}

func TestInviteLinks_Get(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	link := "https://gotd.dev"
	ch := m.Chat(getTestChat())
	links := ch.InviteLinks()

	mock.ExpectCall(&tg.MessagesGetExportedChatInviteRequest{
		Peer: ch.InputPeer(),
		Link: link,
	}).ThenRPCErr(getTestError())
	_, err := links.Get(ctx, link)
	a.Error(err)

	testExported := testChatInviteExported()
	replacer := testExported
	replacer.Link += "/aboba"
	result := &tg.MessagesExportedChatInviteReplaced{
		Invite:    testExported,
		NewInvite: replacer,
		Users:     []tg.UserClass{getTestUser()},
	}
	mock.ExpectCall(&tg.MessagesGetExportedChatInviteRequest{
		Peer: ch.InputPeer(),
		Link: link,
	}).ThenResult(result)
	got, err := links.Get(ctx, link)
	a.NoError(err)
	a.Equal(result.Invite, got.Raw())

	u, err := got.Creator(ctx)
	a.NoError(err)
	a.Equal(testExported.AdminID, u.ID())
}

func TestInviteLinks_edit(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	link := "https://gotd.dev"
	opts := testExportLinkOptions()
	ch := m.Chat(getTestChat())
	links := ch.InviteLinks()

	req := &tg.MessagesEditExportedChatInviteRequest{
		Revoked:       false,
		Peer:          ch.InputPeer(),
		Link:          link,
		ExpireDate:    int(opts.ExpireDate.Unix()),
		UsageLimit:    opts.UsageLimit,
		RequestNeeded: opts.RequestNeeded,
		Title:         opts.Title,
	}
	mock.ExpectCall(req).ThenRPCErr(getTestError())
	_, err := links.Edit(ctx, link, opts)
	a.Error(err)

	result := &tg.MessagesExportedChatInvite{
		Invite: testChatInviteExported(),
	}
	mock.ExpectCall(req).ThenResult(result)
	got, err := links.Edit(ctx, link, opts)
	a.NoError(err)
	a.Equal(result.Invite, got.Raw())

	mock.ExpectCall(&tg.MessagesEditExportedChatInviteRequest{
		Revoked: true,
		Peer:    ch.InputPeer(),
		Link:    link,
	}).ThenResult(result)
	got, err = links.Revoke(ctx, link)
	a.NoError(err)
	a.Equal(result.Invite, got.Raw())
}

func TestInviteLinks_Delete(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	link := "https://gotd.dev"
	ch := m.Chat(getTestChat())
	links := ch.InviteLinks()

	mock.ExpectCall(&tg.MessagesDeleteExportedChatInviteRequest{
		Peer: ch.InputPeer(),
		Link: link,
	}).ThenRPCErr(getTestError())
	a.Error(links.Delete(ctx, link))

	mock.ExpectCall(&tg.MessagesDeleteExportedChatInviteRequest{
		Peer: ch.InputPeer(),
		Link: link,
	}).ThenTrue()
	a.NoError(links.Delete(ctx, link))
}

func TestInviteLinks_hideJoinRequest(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	user := getTestUser().AsInput()
	ch := m.Chat(getTestChat())
	links := ch.InviteLinks()

	mock.ExpectCall(&tg.MessagesHideChatJoinRequestRequest{
		Approved: true,
		Peer:     ch.InputPeer(),
		UserID:   user,
	}).ThenRPCErr(getTestError())
	a.Error(links.ApproveJoin(ctx, user))

	mock.ExpectCall(&tg.MessagesHideChatJoinRequestRequest{
		Approved: false,
		Peer:     ch.InputPeer(),
		UserID:   user,
	}).ThenResult(&tg.Updates{})
	a.NoError(links.DeclineJoin(ctx, user))
}
