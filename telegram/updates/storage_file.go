package updates

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-faster/errors"
)

var (
	_ StateStorage        = (*FileStorage)(nil)
	_ ChannelAccessHasher = (*FileStorage)(nil)
	_ UserAccessHasher    = (*FileStorage)(nil)
)

// fileStorageVersion is the version of the FileStorage format.
const fileStorageVersion = 1

// FileStorage is a StateStorage, ChannelAccessHasher and UserAccessHasher that
// keeps the update state of every user, the pts of their channels and the
// channel and user access hashes in one JSON file, so that a restarted client gets the updates it
// missed while stopped.
//
// Every change rewrites the file atomically (a temporary file in the same
// directory, then rename), so the file is never left half-written, but a
// crash of the machine may lose the last changes and lead to updates being
// delivered again. It suits bots and small accounts; for accounts in many
// active channels use a database-backed storage, e.g. from gotd/contrib.
//
// Use the same FileStorage as Config.Storage and Config.AccessHasher: channel
// state is restored on start only together with its access hash.
//
// A FileStorage must not be shared by processes. It is safe for concurrent
// use.
type FileStorage struct {
	path string

	mux    sync.Mutex
	loaded bool
	data   fileStorageData
}

type fileStorageData struct {
	Version int                        `json:"version"`
	Users   map[int64]*fileStorageUser `json:"users"`
}

type fileStorageUser struct {
	// State is nil until SetState.
	State *State `json:"state,omitempty"`
	// Channels is channel ID to pts.
	Channels map[int64]int `json:"channels,omitempty"`
	// AccessHashes is channel ID to access hash.
	AccessHashes map[int64]int64 `json:"access_hashes,omitempty"`
	// UserAccessHashes is user ID to access hash (added in v0.165.0, still
	// format version 1: older versions ignore it).
	UserAccessHashes map[int64]int64 `json:"user_access_hashes,omitempty"`
}

// NewFileStorage returns a FileStorage that keeps the state in the file at path.
// The file is created on the first change, and its directory must exist.
func NewFileStorage(path string) *FileStorage {
	return &FileStorage{path: path}
}

// load reads the file once. Must be called with s.mux held.
func (s *FileStorage) load() error {
	if s.loaded {
		return nil
	}
	s.data = fileStorageData{Version: fileStorageVersion, Users: map[int64]*fileStorageUser{}}
	raw, err := os.ReadFile(s.path)
	switch {
	case os.IsNotExist(err):
		s.loaded = true
		return nil
	case err != nil:
		return errors.Wrap(err, "read")
	}
	var data fileStorageData
	if err := json.Unmarshal(raw, &data); err != nil {
		return errors.Wrap(err, "decode")
	}
	if data.Version != fileStorageVersion {
		return errors.Errorf("unsupported version %d", data.Version)
	}
	if data.Users == nil {
		data.Users = map[int64]*fileStorageUser{}
	}
	s.data = data
	s.loaded = true
	return nil
}

// save writes the file atomically. Must be called with s.mux held.
//
// On error the changes that were not written are discarded: the file is read
// again by the next call, so the storage never reports a state it did not
// persist.
func (s *FileStorage) save() error {
	if err := s.write(); err != nil {
		s.loaded = false
		return err
	}
	return nil
}

func (s *FileStorage) write() error {
	raw, err := json.Marshal(s.data)
	if err != nil {
		return errors.Wrap(err, "encode")
	}
	f, err := os.CreateTemp(filepath.Dir(s.path), filepath.Base(s.path)+".tmp*")
	if err != nil {
		return errors.Wrap(err, "create temporary file")
	}
	tmp := f.Name()
	if _, err := f.Write(raw); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return errors.Wrap(err, "write")
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return errors.Wrap(err, "close")
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return errors.Wrap(err, "rename")
	}
	return nil
}

// user returns the stored user, nil if absent. Must be called with s.mux held.
func (s *FileStorage) user(userID int64) (*fileStorageUser, error) {
	if err := s.load(); err != nil {
		return nil, errors.Wrap(err, "load")
	}
	return s.data.Users[userID], nil
}

// update calls f with the user that has a state and saves the file, or
// returns an error if the user has no state.
func (s *FileStorage) update(userID int64, f func(u *fileStorageUser)) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	u, err := s.user(userID)
	if err != nil {
		return err
	}
	if u == nil || u.State == nil {
		return errors.New("internalState not found")
	}
	f(u)
	return s.save()
}

// GetState implements StateStorage.
func (s *FileStorage) GetState(_ context.Context, userID int64) (State, bool, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	u, err := s.user(userID)
	if err != nil {
		return State{}, false, err
	}
	if u == nil || u.State == nil {
		return State{}, false, nil
	}
	return *u.State, true, nil
}

// SetState implements StateStorage. It forgets the pts of the user's channels,
// but keeps their access hashes.
func (s *FileStorage) SetState(_ context.Context, userID int64, state State) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	u, err := s.user(userID)
	if err != nil {
		return err
	}
	if u == nil {
		u = &fileStorageUser{}
		s.data.Users[userID] = u
	}
	u.State = &state
	u.Channels = nil
	return s.save()
}

// SetPts implements StateStorage.
func (s *FileStorage) SetPts(_ context.Context, userID int64, pts int) error {
	return s.update(userID, func(u *fileStorageUser) { u.State.Pts = pts })
}

// SetQts implements StateStorage.
func (s *FileStorage) SetQts(_ context.Context, userID int64, qts int) error {
	return s.update(userID, func(u *fileStorageUser) { u.State.Qts = qts })
}

// SetDate implements StateStorage.
func (s *FileStorage) SetDate(_ context.Context, userID int64, date int) error {
	return s.update(userID, func(u *fileStorageUser) { u.State.Date = date })
}

// SetSeq implements StateStorage.
func (s *FileStorage) SetSeq(_ context.Context, userID int64, seq int) error {
	return s.update(userID, func(u *fileStorageUser) { u.State.Seq = seq })
}

// SetDateSeq implements StateStorage.
func (s *FileStorage) SetDateSeq(_ context.Context, userID int64, date, seq int) error {
	return s.update(userID, func(u *fileStorageUser) {
		u.State.Date = date
		u.State.Seq = seq
	})
}

// GetChannelPts implements StateStorage.
func (s *FileStorage) GetChannelPts(_ context.Context, userID, channelID int64) (int, bool, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	u, err := s.user(userID)
	if err != nil {
		return 0, false, err
	}
	if u == nil || u.State == nil {
		return 0, false, nil
	}
	pts, ok := u.Channels[channelID]
	return pts, ok, nil
}

// SetChannelPts implements StateStorage.
func (s *FileStorage) SetChannelPts(_ context.Context, userID, channelID int64, pts int) error {
	return s.update(userID, func(u *fileStorageUser) {
		if u.Channels == nil {
			u.Channels = map[int64]int{}
		}
		u.Channels[channelID] = pts
	})
}

// ForEachChannels implements StateStorage.
func (s *FileStorage) ForEachChannels(
	ctx context.Context, userID int64,
	f func(ctx context.Context, channelID int64, pts int) error,
) error {
	s.mux.Lock()
	u, err := s.user(userID)
	if err != nil {
		s.mux.Unlock()
		return err
	}
	if u == nil || u.State == nil {
		s.mux.Unlock()
		return errors.New("channels map does not exist")
	}
	// Copy, so f may call the storage.
	channels := make(map[int64]int, len(u.Channels))
	for id, pts := range u.Channels {
		channels[id] = pts
	}
	s.mux.Unlock()

	for id, pts := range channels {
		if err := f(ctx, id, pts); err != nil {
			return err
		}
	}
	return nil
}

// GetChannelAccessHash implements ChannelAccessHasher.
func (s *FileStorage) GetChannelAccessHash(_ context.Context, userID, channelID int64) (int64, bool, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	u, err := s.user(userID)
	if err != nil {
		return 0, false, err
	}
	if u == nil {
		return 0, false, nil
	}
	hash, ok := u.AccessHashes[channelID]
	return hash, ok, nil
}

// SetChannelAccessHash implements ChannelAccessHasher.
func (s *FileStorage) SetChannelAccessHash(_ context.Context, userID, channelID, accessHash int64) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	u, err := s.user(userID)
	if err != nil {
		return err
	}
	if u == nil {
		u = &fileStorageUser{}
		s.data.Users[userID] = u
	}
	if old, ok := u.AccessHashes[channelID]; ok && old == accessHash {
		return nil
	}
	if u.AccessHashes == nil {
		u.AccessHashes = map[int64]int64{}
	}
	u.AccessHashes[channelID] = accessHash
	return s.save()
}

// GetUserAccessHash implements UserAccessHasher.
func (s *FileStorage) GetUserAccessHash(_ context.Context, userID, targetUserID int64) (int64, bool, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	u, err := s.user(userID)
	if err != nil {
		return 0, false, err
	}
	if u == nil {
		return 0, false, nil
	}
	hash, ok := u.UserAccessHashes[targetUserID]
	return hash, ok, nil
}

// SetUserAccessHash implements UserAccessHasher.
func (s *FileStorage) SetUserAccessHash(_ context.Context, userID, targetUserID, accessHash int64) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	u, err := s.user(userID)
	if err != nil {
		return err
	}
	if u == nil {
		u = &fileStorageUser{}
		s.data.Users[userID] = u
	}
	if old, ok := u.UserAccessHashes[targetUserID]; ok && old == accessHash {
		return nil
	}
	if u.UserAccessHashes == nil {
		u.UserAccessHashes = map[int64]int64{}
	}
	u.UserAccessHashes[targetUserID] = accessHash
	return s.save()
}
