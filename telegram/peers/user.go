package peers

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
)

// User is user peer.
type User struct {
	raw *tg.User
	m   *Manager
}

// User creates new User, attached to this manager.
func (m *Manager) User(u *tg.User) User {
	m.needsUpdate(userPeerID(u.ID))
	return User{
		raw: u,
		m:   m,
	}
}

// GetUser gets User using given tg.InputUserClass.
func (m *Manager) GetUser(ctx context.Context, p tg.InputUserClass) (User, error) {
	user, err := m.getUser(ctx, p)
	if err != nil {
		return User{}, err
	}
	return m.User(user), nil
}

// Raw returns raw *tg.User.
func (u User) Raw() *tg.User {
	return u.raw
}

// ID returns entity ID.
func (u User) ID() int64 {
	return u.raw.GetID()
}

// TDLibPeerID returns TDLibPeerID for this entity.
func (u User) TDLibPeerID() constant.TDLibPeerID {
	return userPeerID(u.raw.GetID())
}

// VisibleName returns visible name of peer.
//
// It returns FirstName + " " + LastName for users, and title for chats and channels.
func (u User) VisibleName() string {
	firstName := u.raw.FirstName
	lastName := u.raw.LastName
	if lastName == "" {
		return firstName
	}
	return fmt.Sprintf("%s %s", firstName, lastName)
}

// Username returns peer username, if any.
func (u User) Username() (string, bool) {
	return u.raw.GetUsername()
}

// Restricted whether this user/chat/channel is restricted.
func (u User) Restricted() ([]tg.RestrictionReason, bool) {
	reason, ok := u.raw.GetRestrictionReason()
	return reason, ok || u.raw.GetRestricted()
}

// Verified whether this user/chat/channel is verified by Telegram.
func (u User) Verified() bool {
	return u.raw.Verified
}

// Scam whether this user/chat/channel is probably a scam.
func (u User) Scam() bool {
	return u.raw.Scam
}

// Fake whether this user/chat/channel was reported by many users as a fake or scam: be
// careful when interacting with it.
func (u User) Fake() bool {
	return u.raw.Fake
}

// InputPeer returns input peer for this peer.
func (u User) InputPeer() tg.InputPeerClass {
	if u.Self() {
		return &tg.InputPeerSelf{}
	}
	return &tg.InputPeerUser{
		UserID:     u.raw.ID,
		AccessHash: u.raw.AccessHash,
	}
}

// Sync updates current object.
func (u User) Sync(ctx context.Context) error {
	raw, err := u.m.updateUser(ctx, u.InputUser())
	if err != nil {
		return errors.Wrap(err, "get user")
	}
	*u.raw = *raw
	return nil
}

// Manager returns attached Manager.
func (u User) Manager() *Manager {
	return u.m
}

// Report reports a peer for violation of telegram's Terms of Service.
func (u User) Report(ctx context.Context, reason tg.ReportReasonClass, message string) error {
	if _, err := u.m.api.AccountReportPeer(ctx, &tg.AccountReportPeerRequest{
		Peer:    u.InputPeer(),
		Reason:  reason,
		Message: message,
	}); err != nil {
		return errors.Wrap(err, "report")
	}
	return nil
}

// Photo returns peer photo, if any.
func (u User) Photo(ctx context.Context) (*tg.Photo, bool, error) {
	r, err := u.m.api.PhotosGetUserPhotos(ctx, &tg.PhotosGetUserPhotosRequest{
		UserID: u.InputUser(),
		Offset: 0,
		Limit:  1,
	})
	if err != nil {
		return nil, false, errors.Wrap(err, "get user photos")
	}

	if err := u.m.applyUsers(ctx, r.GetUsers()...); err != nil {
		return nil, false, errors.Wrap(err, "apply users")
	}

	photos := r.GetPhotos()
	if len(photos) < 1 {
		return nil, false, nil
	}
	p, ok := photos[0].AsNotEmpty()
	return p, ok, nil
}

// FullRaw returns *tg.UserFull for this User.
func (u User) FullRaw(ctx context.Context) (*tg.UserFull, error) {
	return u.m.getUserFull(ctx, u.InputUser())
}

// ToBot tries to convert this User to Bot.
func (u User) ToBot() (Bot, bool) {
	if !u.raw.Bot {
		return Bot{}, false
	}
	return Bot{
		User: u,
	}, true
}

// Self whether this user indicates the currently logged-in user.
func (u User) Self() bool {
	// TODO(tdakkota): return helper instead?
	return u.raw.Self
}

// Contact whether this user is a contact.
func (u User) Contact() bool {
	return u.raw.Contact
}

// MutualContact whether this user is a mutual contact.
func (u User) MutualContact() bool {
	return u.raw.MutualContact
}

// Deleted whether the account of this user was deleted.
func (u User) Deleted() bool {
	return u.raw.Deleted
}

// Support whether this is an official support user.
func (u User) Support() bool {
	return u.raw.Support
}

// FirstName returns first name.
func (u User) FirstName() (string, bool) {
	return u.raw.GetFirstName()
}

// LastName returns last name.
func (u User) LastName() (string, bool) {
	return u.raw.GetLastName()
}

// Phone returns phone, if any.
func (u User) Phone() (string, bool) {
	return u.raw.GetPhone()
}

// Status returns user status, if any.
func (u User) Status() (tg.UserStatusClass, bool) {
	return u.raw.GetStatus()
}

// LangCode returns users lang code, if any.
func (u User) LangCode() (string, bool) {
	return u.raw.GetLangCode()
}

// InputUser returns input user for this user.
func (u User) InputUser() tg.InputUserClass {
	if u.Self() {
		return &tg.InputUserSelf{}
	}
	return &tg.InputUser{
		UserID:     u.raw.ID,
		AccessHash: u.raw.AccessHash,
	}
}

// ReportSpam reports a new incoming chat for spam, if the peer settings of the chat allow us to do that.
func (u User) ReportSpam(ctx context.Context) error {
	if _, err := u.m.api.MessagesReportSpam(ctx, u.InputPeer()); err != nil {
		return errors.Wrap(err, "report spam")
	}
	return nil
}

// Block blocks this user.
func (u User) Block(ctx context.Context) error {
	if _, err := u.m.api.ContactsBlock(ctx, u.InputPeer()); err != nil {
		return errors.Wrap(err, "block")
	}
	return nil
}

// Unblock unblocks this user.
func (u User) Unblock(ctx context.Context) error {
	if _, err := u.m.api.ContactsUnblock(ctx, u.InputPeer()); err != nil {
		return errors.Wrap(err, "unblock")
	}
	return nil
}

// InviteTo invites User to given channel.
func (u User) InviteTo(ctx context.Context, ch tg.InputChannelClass) error {
	if _, err := u.m.api.ChannelsInviteToChannel(ctx, &tg.ChannelsInviteToChannelRequest{
		Channel: ch,
		Users:   []tg.InputUserClass{u.InputUser()},
	}); err != nil {
		return errors.Wrap(err, "invite to channel")
	}

	return nil
}

// TODO(tdakkota): add more getters, helpers and convertors
