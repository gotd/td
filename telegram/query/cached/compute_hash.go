package cached

import (
	"sort"

	"github.com/gotd/td/telegram/query/hasher"
	"github.com/gotd/td/tg"
)

func (s *ContactsGetContacts) computeHash(v *tg.ContactsContacts) int64 {
	cts := v.Contacts

	sort.SliceStable(cts, func(i, j int) bool {
		return cts[i].UserID < cts[j].UserID
	})
	h := hasher.Hasher{}
	for _, contact := range cts {
		h.Update(uint32(contact.UserID))
	}

	return h.Sum()
}

func (s *MessagesGetQuickReplies) computeHash(v *tg.MessagesQuickReplies) int64 {
	r := v.QuickReplies

	sort.SliceStable(r, func(i, j int) bool {
		return r[i].ShortcutID < r[j].ShortcutID
	})
	h := hasher.Hasher{}
	for _, contact := range r {
		h.Update(uint32(contact.ShortcutID))
	}

	return h.Sum()
}
