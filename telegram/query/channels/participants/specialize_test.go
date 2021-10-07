package participants

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/telegram/message/peer"
	"github.com/nnqq/td/tg"
)

func TestElem(t *testing.T) {
	entities := peer.NewEntities(
		map[int64]*tg.User{10: {}},
		map[int64]*tg.Chat{},
		map[int64]*tg.Channel{},
	)

	type results struct {
		admin, creator, photos bool
	}
	tests := []struct {
		Name string
		Part tg.ChannelParticipantClass
		results
	}{
		{"UnknownPlain", &tg.ChannelParticipant{UserID: 45}, results{}},
		{"UnknownBanned", &tg.ChannelParticipantBanned{Peer: &tg.PeerUser{UserID: 45}},
			results{}},
		{"UnknownAdmin", &tg.ChannelParticipantAdmin{UserID: 45}, results{}},
		{"UnknownCreator", &tg.ChannelParticipantCreator{UserID: 45}, results{}},
		{"Plain", &tg.ChannelParticipant{UserID: 10}, results{photos: true}},
		{"Banned", &tg.ChannelParticipantBanned{Peer: &tg.PeerUser{UserID: 10}},
			results{photos: true}},
		{"Admin", &tg.ChannelParticipantAdmin{UserID: 10}, results{
			admin:  true,
			photos: true,
		}},
		{"Creator", &tg.ChannelParticipantCreator{UserID: 10}, results{
			creator: true,
			photos:  true,
		}},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			a := require.New(t)
			var ok bool

			elem := Elem{Participant: test.Part, Entities: entities}
			_, ok = elem.UserPhotos(nil)
			a.Equal(test.photos, ok)
			_, _, ok = elem.Admin()
			a.Equal(test.admin, ok)
			_, _, ok = elem.Creator()
			a.Equal(test.creator, ok)
		})
	}
}
