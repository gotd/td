package peers

import (
	"context"
	"sync"

	"go.uber.org/atomic"

	"github.com/gotd/td/tg"
)

// InmemoryStorage is basic in-memory Storage implementation.
type InmemoryStorage struct {
	phones  map[string]Key
	data    map[Key]Value
	dataMux sync.Mutex // guards phones and data

	contactsHash atomic.Int64
}

var _ Storage = (*InmemoryStorage)(nil)

func (f *InmemoryStorage) initLocked() {
	if f.phones == nil {
		f.phones = map[string]Key{}
	}
	if f.data == nil {
		f.data = map[Key]Value{}
	}
}

// Save implements Storage.
func (f *InmemoryStorage) Save(ctx context.Context, key Key, value Value) error {
	f.dataMux.Lock()
	defer f.dataMux.Unlock()
	f.initLocked()

	f.data[key] = value
	return nil
}

// Find implements Storage.
func (f *InmemoryStorage) Find(ctx context.Context, key Key) (value Value, found bool, _ error) {
	f.dataMux.Lock()
	defer f.dataMux.Unlock()

	value, found = f.data[key]
	return value, found, nil
}

// SavePhone implements Storage.
func (f *InmemoryStorage) SavePhone(ctx context.Context, phone string, key Key) error {
	f.dataMux.Lock()
	defer f.dataMux.Unlock()
	f.initLocked()

	f.phones[phone] = key
	return nil
}

// FindPhone implements Storage.
func (f *InmemoryStorage) FindPhone(ctx context.Context, phone string) (key Key, value Value, found bool, err error) {
	f.dataMux.Lock()
	defer f.dataMux.Unlock()

	key, found = f.phones[phone]
	if !found {
		return Key{}, Value{}, false, nil
	}
	value, found = f.data[key]
	return key, value, found, nil
}

// GetContactsHash implements Storage.
func (f *InmemoryStorage) GetContactsHash(ctx context.Context) (int64, error) {
	v := f.contactsHash.Load()
	return v, nil
}

// SaveContactsHash implements Storage.
func (f *InmemoryStorage) SaveContactsHash(ctx context.Context, hash int64) error {
	f.contactsHash.Store(hash)
	return nil
}

// InmemoryCache is basic in-memory Cache implementation.
type InmemoryCache struct {
	users    map[int64]*tg.User
	usersMux sync.Mutex

	usersFull    map[int64]*tg.UserFull
	usersFullMux sync.Mutex

	chats    map[int64]*tg.Chat
	chatsMux sync.Mutex

	chatsFull    map[int64]*tg.ChatFull
	chatsFullMux sync.Mutex

	channels    map[int64]*tg.Channel
	channelsMux sync.Mutex

	channelsFull    map[int64]*tg.ChannelFull
	channelsFullMux sync.Mutex
}

// SaveUsers implements Cache.
func (f *InmemoryCache) SaveUsers(ctx context.Context, users ...*tg.User) error {
	f.usersMux.Lock()
	defer f.usersMux.Unlock()
	if f.channelsFull == nil {
		f.users = map[int64]*tg.User{}
	}

	for _, u := range users {
		f.users[u.GetID()] = u
	}

	return nil
}

// SaveUserFulls implements Cache.
func (f *InmemoryCache) SaveUserFulls(ctx context.Context, users ...*tg.UserFull) error {
	f.usersFullMux.Lock()
	defer f.usersFullMux.Unlock()
	if f.channelsFull == nil {
		f.usersFull = map[int64]*tg.UserFull{}
	}

	for _, u := range users {
		f.usersFull[u.GetID()] = u
	}

	return nil
}

// FindUser implements Cache.
func (f *InmemoryCache) FindUser(ctx context.Context, id int64) (*tg.User, bool, error) {
	f.usersMux.Lock()
	defer f.usersMux.Unlock()

	u, ok := f.users[id]
	return u, ok, nil
}

// FindUserFull implements Cache.
func (f *InmemoryCache) FindUserFull(ctx context.Context, id int64) (*tg.UserFull, bool, error) {
	f.usersFullMux.Lock()
	defer f.usersFullMux.Unlock()

	u, ok := f.usersFull[id]
	return u, ok, nil
}

// SaveChats implements Cache.
func (f *InmemoryCache) SaveChats(ctx context.Context, chats ...*tg.Chat) error {
	f.chatsMux.Lock()
	defer f.chatsMux.Unlock()
	if f.channelsFull == nil {
		f.chats = map[int64]*tg.Chat{}
	}

	for _, c := range chats {
		f.chats[c.GetID()] = c
	}

	return nil
}

// SaveChatFulls implements Cache.
func (f *InmemoryCache) SaveChatFulls(ctx context.Context, chats ...*tg.ChatFull) error {
	f.chatsFullMux.Lock()
	defer f.chatsFullMux.Unlock()
	if f.channelsFull == nil {
		f.chatsFull = map[int64]*tg.ChatFull{}
	}

	for _, c := range chats {
		f.chatsFull[c.GetID()] = c
	}

	return nil
}

// FindChat implements Cache.
func (f *InmemoryCache) FindChat(ctx context.Context, id int64) (*tg.Chat, bool, error) {
	f.chatsMux.Lock()
	defer f.chatsMux.Unlock()

	c, ok := f.chats[id]
	return c, ok, nil
}

// FindChatFull implements Cache.
func (f *InmemoryCache) FindChatFull(ctx context.Context, id int64) (*tg.ChatFull, bool, error) {
	f.chatsFullMux.Lock()
	defer f.chatsFullMux.Unlock()

	c, ok := f.chatsFull[id]
	return c, ok, nil
}

// SaveChannels implements Cache.
func (f *InmemoryCache) SaveChannels(ctx context.Context, channels ...*tg.Channel) error {
	f.channelsMux.Lock()
	defer f.channelsMux.Unlock()
	if f.channelsFull == nil {
		f.channels = map[int64]*tg.Channel{}
	}

	for _, c := range channels {
		f.channels[c.GetID()] = c
	}

	return nil
}

// SaveChannelFulls implements Cache.
func (f *InmemoryCache) SaveChannelFulls(ctx context.Context, channels ...*tg.ChannelFull) error {
	f.channelsFullMux.Lock()
	defer f.channelsFullMux.Unlock()
	if f.channelsFull == nil {
		f.channelsFull = map[int64]*tg.ChannelFull{}
	}

	for _, c := range channels {
		f.channelsFull[c.GetID()] = c
	}

	return nil
}

// FindChannel implements Cache.
func (f *InmemoryCache) FindChannel(ctx context.Context, id int64) (*tg.Channel, bool, error) {
	f.channelsMux.Lock()
	defer f.channelsMux.Unlock()

	c, ok := f.channels[id]
	return c, ok, nil
}

// FindChannelFull implements Cache.
func (f *InmemoryCache) FindChannelFull(ctx context.Context, id int64) (*tg.ChannelFull, bool, error) {
	f.channelsFullMux.Lock()
	defer f.channelsFullMux.Unlock()

	c, ok := f.channelsFull[id]
	return c, ok, nil
}
