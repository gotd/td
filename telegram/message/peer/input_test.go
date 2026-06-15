package peer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

type stubResolver struct {
	peer tg.InputPeerClass
}

func (s stubResolver) ResolveDomain(context.Context, string) (tg.InputPeerClass, error) {
	return s.peer, nil
}

func (s stubResolver) ResolvePhone(context.Context, string) (tg.InputPeerClass, error) {
	return s.peer, nil
}

func TestResolveInputPeer(t *testing.T) {
	ctx := context.Background()
	want := &tg.InputPeerUser{UserID: 1, AccessHash: 2}
	r := stubResolver{peer: want}

	t.Run("Concrete", func(t *testing.T) {
		got, err := ResolveInputPeer(ctx, r, want)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("Resolved", func(t *testing.T) {
		got, err := ResolveInputPeer(ctx, r, Resolve("telegram"))
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("Nil", func(t *testing.T) {
		_, err := ResolveInputPeer(ctx, r, nil)
		require.Error(t, err)
	})

	t.Run("Unsupported", func(t *testing.T) {
		// An InputUser satisfies InputPeer (Zero/String) but is not a peer.
		_, err := ResolveInputPeer(ctx, r, &tg.InputUser{UserID: 1})
		require.Error(t, err)
	})
}

func TestResolvedBind(t *testing.T) {
	ctx := context.Background()
	want := &tg.InputPeerUser{UserID: 1, AccessHash: 2}

	got, err := Resolve("telegram").Bind(stubResolver{peer: want})(ctx)
	require.NoError(t, err)
	require.Equal(t, want, got)

	var empty *Resolved
	require.True(t, empty.Zero())
}
