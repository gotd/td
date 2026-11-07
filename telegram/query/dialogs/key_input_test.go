package dialogs

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestDialogKeyInputPeer(t *testing.T) {
	for _, p := range []tg.InputPeerClass{
		&tg.InputPeerUser{UserID: 1, AccessHash: 2},
		&tg.InputPeerChat{ChatID: 3},
		&tg.InputPeerChannel{ChannelID: 4, AccessHash: 5},
	} {
		var k DialogKey
		require.NoError(t, k.FromInputPeer(p))
		require.Equal(t, p, k.InputPeer())
	}
	require.Nil(t, DialogKey{Kind: PeerKind(42)}.InputPeer())
}
