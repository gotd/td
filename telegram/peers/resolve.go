package peers

import (
	"context"
	"strings"

	"github.com/go-faster/errors"

	"github.com/gotd/td/internal/ascii"
	"github.com/gotd/td/telegram/internal/deeplink"
	"github.com/gotd/td/tg"
)

// Resolve uses given string to create new peer promise.
//
// Input examples:
//
//	@telegram
//	telegram
//	t.me/telegram
//	https://t.me/telegram
//	tg:resolve?domain=telegram
//	tg://resolve?domain=telegram
//	+13115552368
//	+1 (311) 555-0123
//	+1 311 555-6162
//	13115556162
func (m *Manager) Resolve(ctx context.Context, from string) (Peer, error) {
	from = strings.TrimSpace(from)

	if deeplink.IsDeeplinkLike(from) {
		return m.ResolveDeeplink(ctx, from)
	}
	if isPhoneNumber(from) {
		return m.ResolvePhone(ctx, from)
	}
	return m.ResolveDomain(ctx, from)
}

func isPhoneNumber(s string) bool {
	if s == "" {
		return false
	}
	r := rune(s[0])
	return r == '+' || ascii.IsDigit(r)
}

func cleanupPhone(phone string) string {
	var needClean bool
	for _, ch := range phone {
		if !ascii.IsDigit(ch) {
			needClean = true
			break
		}
	}
	if !needClean {
		return phone
	}

	clean := strings.Builder{}
	clean.Grow(len(phone) + 1)

	for _, ch := range phone {
		if ascii.IsDigit(ch) {
			clean.WriteRune(ch)
		}
	}

	return clean.String()
}

// ResolvePhone uses given phone to resolve User.
//
// Input example:
//
//	+13115552368
//	+1 (311) 555-0123
//	+1 311 555-6162
//	13115556162
//
// Note that Telegram represents phone numbers according to the E.164 standard
// without the plus sign (”+”) prefix. The resolver therefore takes an easy
// route and just deletes any non-digit symbols from phone number string.
func (m *Manager) ResolvePhone(ctx context.Context, phone string) (User, error) {
	tried := false

	phone = cleanupPhone(phone)
	for {
		if tried {
			return User{}, &PhoneNotFoundError{Phone: phone}
		}

		key, v, found, err := m.storage.FindPhone(ctx, phone)
		if err != nil {
			return User{}, errors.Wrap(err, "find by phone")
		}
		if found {
			return m.GetUser(ctx, &tg.InputUser{
				UserID:     key.ID,
				AccessHash: v.AccessHash,
			})
		}
		if m.selfIsBot() {
			return User{}, &PhoneNotFoundError{Phone: phone}
		}
		tried = true

		users, err := m.updateContacts(ctx)
		if err != nil {
			return User{}, errors.Wrap(err, "update contacts")
		}

		for _, user := range users {
			if u, ok := user.AsNotEmpty(); ok && u.Phone == phone {
				return m.User(u), nil
			}
		}
	}
}

func validateDomain(domain string) error {
	return deeplink.ValidateDomain(domain)
}

func (m *Manager) findPeerClass(p tg.PeerClass, users []tg.UserClass, chats []tg.ChatClass) (Peer, bool) {
	switch p := p.(type) {
	case *tg.PeerUser:
		for _, user := range users {
			u, ok := user.AsNotEmpty()
			if ok && u.ID == p.UserID {
				return m.User(u), true
			}
		}
	case *tg.PeerChat:
		for _, chat := range chats {
			c, ok := chat.(*tg.Chat)
			if ok && c.ID == p.ChatID {
				return m.Chat(c), true
			}
		}
	case *tg.PeerChannel:
		for _, chat := range chats {
			c, ok := chat.(*tg.Channel)
			if ok && c.ID == p.ChannelID {
				return m.Channel(c), true
			}
		}
	}
	return nil, false
}

// ResolveDomain uses given domain to create new peer promise.
//
// May be prefixed with @ or not.
//
// Input examples:
//
//	@telegram
//	telegram
func (m *Manager) ResolveDomain(ctx context.Context, domain string) (Peer, error) {
	domain = strings.TrimPrefix(domain, "@")

	if err := validateDomain(domain); err != nil {
		return nil, errors.Wrap(err, "validate domain")
	}

	ch := m.sg.DoChan(domain, func() (interface{}, error) {
		result, err := m.api.ContactsResolveUsername(ctx, domain)
		if err != nil {
			return nil, errors.Wrap(err, "resolve")
		}
		return result, nil
	})

	var result *tg.ContactsResolvedPeer
	select {
	case r := <-ch:
		if err := r.Err; err != nil {
			return nil, r.Err
		}
		result = r.Val.(*tg.ContactsResolvedPeer)
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	if err := m.applyEntities(ctx, result.Users, result.Chats); err != nil {
		return nil, err
	}

	p, ok := m.findPeerClass(result.Peer, result.Users, result.Chats)
	if !ok {
		return nil, &PeerNotFoundError{Peer: result.Peer}
	}

	return p, nil
}

// ResolveDeeplink uses given deeplink to create new peer promise.
//
// Input examples:
//
//	t.me/telegram
//	https://t.me/telegram
//	tg:resolve?domain=telegram
//	tg://resolve?domain=telegram
func (m *Manager) ResolveDeeplink(ctx context.Context, u string) (Peer, error) {
	link, err := deeplink.Expect(u, deeplink.Resolve)
	if err != nil {
		return nil, err
	}
	domain := link.Args.Get("domain")

	if err := validateDomain(domain); err != nil {
		return nil, errors.Wrap(err, "validate domain")
	}

	return m.ResolveDomain(ctx, domain)
}
