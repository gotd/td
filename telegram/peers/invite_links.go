package peers

import (
	"context"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// InviteLinks represents invite links of Chat or Channel.
type InviteLinks struct {
	peer Peer
	m    *Manager
}

// ExportLinkOptions is options for ExportNew.
type ExportLinkOptions struct {
	// Whether users joining the chat via the link need to be approved by chat administrators.
	RequestNeeded bool
	// Expiration date.
	//
	// If zero, will not be used.
	ExpireDate time.Time
	// Maximum number of users that can join using this link.
	//
	// If zero, will not be used.
	UsageLimit int
	// Title of this link.
	//
	// If zero, will not be used.
	Title string
}

// ExportNew creates new primary invite link for a chat.
//
// Notice: Any previously generated primary link is revoked.
//
// See also AddNew.
func (e InviteLinks) ExportNew(ctx context.Context, opts ExportLinkOptions) (InviteLink, error) {
	return e.newLink(ctx, true, opts)
}

// AddNew creates an additional invite link for a chat.
func (e InviteLinks) AddNew(ctx context.Context, opts ExportLinkOptions) (InviteLink, error) {
	return e.newLink(ctx, false, opts)
}

// Get returns link info.
func (e InviteLinks) Get(ctx context.Context, link string) (InviteLink, error) {
	r, err := e.m.api.MessagesGetExportedChatInvite(ctx, &tg.MessagesGetExportedChatInviteRequest{
		Peer: e.peer.InputPeer(),
		Link: link,
	})
	if err != nil {
		return InviteLink{}, errors.Wrap(err, "get chat invite")
	}

	return e.applyExportedInvite(ctx, r)
}

// Edit edits link info.
func (e InviteLinks) Edit(ctx context.Context, link string, opts ExportLinkOptions) (InviteLink, error) {
	req := tg.MessagesEditExportedChatInviteRequest{
		Revoked:       false,
		Peer:          e.peer.InputPeer(),
		Link:          link,
		ExpireDate:    0,
		UsageLimit:    opts.UsageLimit,
		RequestNeeded: opts.RequestNeeded,
		Title:         opts.Title,
	}
	if e := opts.ExpireDate; !e.IsZero() {
		req.ExpireDate = int(e.Unix())
	}
	return e.edit(ctx, "edit chat invite", req)
}

// Revoke revokes invite link and returns revoked link info.
//
// If the primary link is revoked, a new link is automatically generated.
func (e InviteLinks) Revoke(ctx context.Context, link string) (InviteLink, error) {
	return e.edit(ctx, "revoke chat invite", tg.MessagesEditExportedChatInviteRequest{
		Revoked: true,
		Peer:    e.peer.InputPeer(),
		Link:    link,
	})
}

func (e InviteLinks) edit(
	ctx context.Context,
	msg string,
	req tg.MessagesEditExportedChatInviteRequest,
) (InviteLink, error) {
	r, err := e.m.api.MessagesEditExportedChatInvite(ctx, &req)
	if err != nil {
		return InviteLink{}, errors.Wrap(err, msg)
	}

	return e.applyExportedInvite(ctx, r)
}

// Delete deletes invite link.
//
// Not available for bots.
func (e InviteLinks) Delete(ctx context.Context, link string) error {
	if _, err := e.m.api.MessagesDeleteExportedChatInvite(ctx, &tg.MessagesDeleteExportedChatInviteRequest{
		Peer: e.peer.InputPeer(),
		Link: link,
	}); err != nil {
		return errors.Wrap(err, "delete chat invite")
	}

	return nil
}

func (e InviteLinks) applyExportedInvite(
	ctx context.Context,
	r tg.MessagesExportedChatInviteClass,
) (InviteLink, error) {
	if err := e.m.applyUsers(ctx, r.GetUsers()...); err != nil {
		return InviteLink{}, errors.Wrap(err, "update users")
	}

	switch r := r.(type) {
	case *tg.MessagesExportedChatInviteReplaced:
		return e.replacedLink(r.GetInvite(), r.GetNewInvite()), nil
	}

	return e.inviteLink(r.GetInvite()), nil
}

func (e InviteLinks) newLink(
	ctx context.Context,
	revokeOld bool,
	opts ExportLinkOptions,
) (InviteLink, error) {
	req := &tg.MessagesExportChatInviteRequest{
		LegacyRevokePermanent: revokeOld,
		RequestNeeded:         opts.RequestNeeded,
		Peer:                  e.peer.InputPeer(),
		ExpireDate:            0,
		UsageLimit:            opts.UsageLimit,
		Title:                 opts.Title,
	}
	if e := opts.ExpireDate; !e.IsZero() {
		req.ExpireDate = int(e.Unix())
	}

	invite, err := e.m.api.MessagesExportChatInvite(ctx, req)
	if err != nil {
		return InviteLink{}, errors.Wrap(err, "create invite")
	}
	return e.inviteLink(*invite), nil
}

// TODO(tdakkota): add methods with pagination, when query will be updated
