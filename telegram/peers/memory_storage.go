package peers

import (
	"context"
	"sync"

	"go.uber.org/atomic"
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
