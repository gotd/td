package cached

import (
	"sort"

	"github.com/nnqq/td/telegram/query/hasher"
	"github.com/nnqq/td/tg"
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
