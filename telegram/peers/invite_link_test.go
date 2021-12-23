package peers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInviteLinkGetters(t *testing.T) {
	a := require.New(t)

	testExported := testChatInviteExported()
	replacer := testExported
	replacer.Link += "/aboba"
	link := InviteLink{
		raw:       testExported,
		newInvite: replacer,
	}
	a.Equal(testExported, link.Raw())
	{
		replacedWith, ok := link.ReplacedWith()
		a.True(ok)
		a.Equal(replacer, replacedWith.Raw())
	}
	a.Equal(testExported.Revoked, link.Revoked())
	a.Equal(testExported.Permanent, link.Permanent())
	a.Equal(testExported.RequestNeeded, link.RequestNeeded())
	a.Equal(testExported.Link, link.Link())
	{
		date, _ := link.ExpireDate()
		a.Equal(testExported.ExpireDate, int(date.Unix()))
	}
}
