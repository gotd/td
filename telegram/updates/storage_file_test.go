package updates

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileStorage(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "updates.json")
	const user = 10

	s := NewFileStorage(path)
	_, found, err := s.GetState(ctx, user)
	require.NoError(t, err)
	require.False(t, found)
	require.Error(t, s.SetPts(ctx, user, 1), "no state yet")
	require.Error(t, s.SetChannelPts(ctx, user, 5, 1), "no state yet")
	require.Error(t, s.ForEachChannels(ctx, user, func(context.Context, int64, int) error { return nil }))
	_, err = os.Stat(path)
	require.True(t, os.IsNotExist(err), "nothing written before a change")

	require.NoError(t, s.SetChannelAccessHash(ctx, user, 5, 55))
	require.NoError(t, s.SetState(ctx, user, State{Pts: 1, Qts: 2, Date: 3, Seq: 4}))
	require.NoError(t, s.SetPts(ctx, user, 11))
	require.NoError(t, s.SetQts(ctx, user, 12))
	require.NoError(t, s.SetDate(ctx, user, 13))
	require.NoError(t, s.SetSeq(ctx, user, 14))
	require.NoError(t, s.SetDateSeq(ctx, user, 23, 24))
	require.NoError(t, s.SetChannelPts(ctx, user, 5, 100))
	require.NoError(t, s.SetChannelPts(ctx, user, 6, 200))

	// A new storage reads everything back from the file.
	s2 := NewFileStorage(path)
	state, found, err := s2.GetState(ctx, user)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, State{Pts: 11, Qts: 12, Date: 23, Seq: 24}, state)
	pts, found, err := s2.GetChannelPts(ctx, user, 5)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, 100, pts)
	hash, found, err := s2.GetChannelAccessHash(ctx, user, 5)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(55), hash)
	channels := map[int64]int{}
	require.NoError(t, s2.ForEachChannels(ctx, user, func(_ context.Context, id int64, pts int) error {
		channels[id] = pts
		// The storage may be called back.
		return s2.SetChannelPts(ctx, user, id, pts+1)
	}))
	require.Equal(t, map[int64]int{5: 100, 6: 200}, channels)

	// SetState forgets channel pts, but keeps access hashes.
	require.NoError(t, s2.SetState(ctx, user, State{Pts: 1}))
	_, found, err = s2.GetChannelPts(ctx, user, 5)
	require.NoError(t, err)
	require.False(t, found)
	_, found, err = s2.GetChannelAccessHash(ctx, user, 5)
	require.NoError(t, err)
	require.True(t, found)

	// Other users are separate.
	_, found, err = s2.GetState(ctx, user+1)
	require.NoError(t, err)
	require.False(t, found)

	matches, err := filepath.Glob(filepath.Join(filepath.Dir(path), "*.tmp*"))
	require.NoError(t, err)
	require.Empty(t, matches, "temporary files must be removed")
}

func TestFileStorageWriteError(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "updates.json")
	s := NewFileStorage(path)
	require.NoError(t, s.SetState(ctx, 1, State{Pts: 1}))

	// Make the directory read-only, so that writes fail.
	require.NoError(t, os.Chmod(dir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
	if f, err := os.CreateTemp(dir, "probe"); err == nil {
		// Running as root: permissions are not enforced.
		_ = f.Close()
		_ = os.Remove(f.Name())
		t.Skip("directory permissions are not enforced")
	}
	require.Error(t, s.SetPts(ctx, 1, 2))

	// The change that was not written is not reported.
	state, found, err := s.GetState(ctx, 1)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, 1, state.Pts)
}

func TestFileStorageCorrupted(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "updates.json")
	require.NoError(t, os.WriteFile(path, []byte("{"), 0o600))
	_, _, err := NewFileStorage(path).GetState(ctx, 1)
	require.Error(t, err)

	require.NoError(t, os.WriteFile(path, []byte(`{"version":99}`), 0o600))
	_, _, err = NewFileStorage(path).GetState(ctx, 1)
	require.ErrorContains(t, err, "unsupported version")
}

func TestFileStorageUserAccessHash(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "updates.json")
	s := NewFileStorage(path)
	_, found, err := s.GetUserAccessHash(ctx, 1, 10)
	require.NoError(t, err)
	require.False(t, found)
	require.NoError(t, s.SetUserAccessHash(ctx, 1, 10, 777))

	// Survives a restart.
	s = NewFileStorage(path)
	hash, found, err := s.GetUserAccessHash(ctx, 1, 10)
	require.NoError(t, err)
	require.True(t, found)
	require.EqualValues(t, 777, hash)
}
