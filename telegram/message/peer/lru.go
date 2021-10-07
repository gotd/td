package peer

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/nnqq/td/clock"
	"github.com/nnqq/td/tg"
)

// LRUResolver is simple decorator for Resolver to cache result in LRU.
type LRUResolver struct {
	next  Resolver
	clock clock.Clock

	expiration time.Duration
	capacity   int

	cache   map[string]*linkedNode
	lruList *linkedList
	// Guards LRU state — cache and lruList
	mux sync.Mutex

	// Prevents multiple identical requests at the same time.
	sg singleflight.Group
}

// NewLRUResolver creates new LRUResolver.
func NewLRUResolver(next Resolver, capacity int) *LRUResolver {
	return &LRUResolver{
		next:       next,
		clock:      clock.System,
		expiration: time.Minute,
		capacity:   capacity,
		cache:      make(map[string]*linkedNode, capacity),
		lruList:    &linkedList{},
		sg:         singleflight.Group{},
	}
}

// WithClock sets clock to use when counting expiration.
func (l *LRUResolver) WithClock(c clock.Clock) *LRUResolver {
	l.clock = c
	return l
}

// WithExpiration sets expiration timeout for records in cache.
// If zero, expiration will be disabled. Default value is a minute.
func (l *LRUResolver) WithExpiration(expiration time.Duration) *LRUResolver {
	l.expiration = expiration
	return l
}

// Evict deletes record from cache.
func (l *LRUResolver) Evict(key string) (tg.InputPeerClass, bool) {
	return l.delete(key)
}

// ResolveDomain implements Resolver.
func (l *LRUResolver) ResolveDomain(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	if v, ok := l.get(domain); ok {
		return v, nil
	}

	r, err := l.next.ResolveDomain(ctx, domain)
	if err != nil {
		return nil, err
	}

	l.put(domain, r)
	return r, nil
}

// ResolvePhone implements Resolver.
func (l *LRUResolver) ResolvePhone(ctx context.Context, phone string) (tg.InputPeerClass, error) {
	if v, ok := l.get(phone); ok {
		return v, nil
	}

	r, err := l.next.ResolvePhone(ctx, phone)
	if err != nil {
		return nil, err
	}

	l.put(phone, r)
	return r, nil
}

func (l *LRUResolver) get(key string) (v tg.InputPeerClass, ok bool) {
	l.mux.Lock()
	defer l.mux.Unlock()

	if found, ok := l.cache[key]; ok {
		// Delete expired and return false.
		if l.expiration > 0 && l.clock.Now().After(found.expiresAt) {
			l.deleteLocked(key)
			return nil, false
		}
		l.lruList.MoveToFront(found)
		return found.value, true
	}
	return
}

func (l *LRUResolver) put(key string, value tg.InputPeerClass) {
	l.mux.Lock()
	defer l.mux.Unlock()

	if found, ok := l.cache[key]; ok {
		found.value = value
		l.lruList.MoveToFront(found)
	} else {
		if len(l.cache) >= l.capacity {
			l.deleteLocked(l.lruList.Back().key)
		}

		l.cache[key] = l.lruList.PushFront(nodeData{
			key,
			value,
			l.clock.Now().Add(l.expiration),
		})
	}
}

func (l *LRUResolver) delete(key string) (tg.InputPeerClass, bool) {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.deleteLocked(key)
}

// deleteLocked deletes record from cache.
// Assumes mutex is locked.
func (l *LRUResolver) deleteLocked(key string) (tg.InputPeerClass, bool) {
	found, ok := l.cache[key]
	if !ok {
		return nil, false
	}

	l.lruList.Remove(found)
	delete(l.cache, key)
	return nil, true
}
