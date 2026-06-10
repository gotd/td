package updates

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func TestSaveUserHashes(t *testing.T) {
	ctx := context.Background()
	handler := telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
		return nil
	})
	s := newShortTestState(t, &countDiffAPI{}, handler)

	// Pre-seed a known user to exercise the "already known, skip" branch.
	require.NoError(t, s.userHasher.SetUserAccessHash(ctx, s.selfID, 111, 1111))

	s.saveUserHashes(ctx, []tg.UserClass{
		&tg.User{ID: 111, AccessHash: 9999},     // already known: must not be overwritten
		&tg.User{ID: 222, AccessHash: 2222},     // full: recorded
		&tg.User{ID: 333, Min: true, AccessHash: 3333}, // min: skipped
		&tg.User{ID: 444, AccessHash: 0},        // zero hash: skipped
		&tg.UserEmpty{ID: 555},                  // not *tg.User: skipped
	})

	assertHash := func(id, want int64, wantFound bool) {
		t.Helper()
		hash, found, err := s.userHasher.GetUserAccessHash(ctx, s.selfID, id)
		require.NoError(t, err)
		require.Equal(t, wantFound, found)
		if wantFound {
			require.Equal(t, want, hash)
		}
	}

	assertHash(111, 1111, true) // unchanged
	assertHash(222, 2222, true) // newly recorded
	assertHash(333, 0, false)   // min skipped
	assertHash(444, 0, false)   // zero hash skipped
	assertHash(555, 0, false)   // wrong type skipped
}
