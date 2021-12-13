package peers

import (
	"context"
	"fmt"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// User is user peer.
type User struct {
	raw *tg.User
	m   *Manager
}

// User creates new User, attached to this manager.
func (m *Manager) User(u *tg.User) User {
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

// VisibleName returns visible name of peer.
//
// It returns FirstName + " " + LastName for users, and title for chats and channels.
func (u User) VisibleName() string {
	firstName, _ := u.raw.GetFirstName()
	lastName, _ := u.raw.GetLastName()
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
	return u.raw.GetVerified()
}

// Scam whether this user/chat/channel is probably a scam.
func (u User) Scam() bool {
	return u.raw.GetScam()
}

// Fake whether this user/chat/channel was reported by many users as a fake or scam: be
// careful when interacting with it.
func (u User) Fake() bool {
	return u.raw.GetFake()
}

// InputPeer returns input peer for this peer.
func (u User) InputPeer() tg.InputPeerClass {
	return u.raw.AsInputPeer()
}

// Sync updates current object.
func (u User) Sync(ctx context.Context) error {
	raw, err := u.m.getUser(ctx, u.InputUser())
	if err != nil {
		return errors.Wrap(err, "get user")
	}
	*u.raw = *raw
	return nil
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

// ToBot tries to convert this User to Bot.
func (u User) ToBot() (Bot, bool) {
	if !u.raw.GetBot() {
		return Bot{}, false
	}
	return Bot{
		User: u,
	}, true
}

// Contact whether this user is a contact.
func (u User) Contact() bool {
	return u.raw.GetContact()
}

// MutualContact whether this user is a mutual contact.
func (u User) MutualContact() bool {
	return u.raw.GetMutualContact()
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
func (u User) InputUser() *tg.InputUser {
	return u.raw.AsInput()
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
