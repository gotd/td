package mtproto

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/crypto"
)

func TestSessionStateReuse(t *testing.T) {
	ctx := context.Background()
	key := crypto.Key{1}.WithID()
	state := NewSessionState()
	require.True(t, state.TryAcquire())
	require.False(t, state.TryAcquire(), "one connection at a time")

	first := New(nil, Options{Key: key, SessionState: state})
	require.False(t, first.reuseSession(ctx, key), "nothing to continue yet")
	require.NoError(t, first.newSessionID())
	first.sentContentMessages = 7
	id := first.session().ID
	first.releaseSessionState()

	require.True(t, state.TryAcquire())
	second := New(nil, Options{Key: key, SessionState: state})
	require.True(t, second.reuseSession(ctx, key))
	require.Equal(t, id, second.session().ID)
	require.Equal(t, int32(7), second.sentContentMessages)
	require.True(t, second.sessionContinued)
	second.releaseSessionState()

	// A different key starts a new session.
	require.True(t, state.TryAcquire())
	other := crypto.Key{2}.WithID()
	third := New(nil, Options{Key: other, SessionState: state})
	require.False(t, third.reuseSession(ctx, other))
	third.releaseSessionState()

	// PFS connections never continue sessions.
	require.True(t, state.TryAcquire())
	pfs := New(nil, Options{PermKey: key, EnablePFS: true, SessionState: state})
	require.False(t, pfs.reuseSession(ctx, key))
}
