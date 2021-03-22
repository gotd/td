package peer

import (
	"context"
	"sync"

	"github.com/gotd/td/tg"
)

// LRUResolver is simple decorator for Resolver
// to cache result in LRU.
type LRUResolver struct {
	next Resolver

	capacity int
	cache    map[string]*linkedNode
	lruList  *linkedList

	mux sync.Mutex
}

// NewLRUResolver creates new LRUResolver.
func NewLRUResolver(next Resolver, capacity int) *LRUResolver {
	return &LRUResolver{
		next:     next,
		capacity: capacity,
		cache:    make(map[string]*linkedNode, capacity),
		lruList:  &linkedList{},
	}
}

// ResolveDomain implements Resolver.
func (l *LRUResolver) ResolveDomain(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	// TODO(tdakkota): expiration support
	// TODO(tdakkota): resolve race conditions in case when two and more goroutines tries to fetch same domain.
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
			l.delete(l.lruList.Back().key)
		}

		l.cache[key] = l.lruList.PushFront(nodeData{key, value})
	}
}

func (l *LRUResolver) delete(key string) bool {
	found, ok := l.cache[key]
	if !ok {
		return false
	}

	l.lruList.Remove(found)
	delete(l.cache, key)
	return true
}
