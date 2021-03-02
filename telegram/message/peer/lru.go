package peer

import (
	"context"

	"github.com/gotd/td/tg"
)

// LRUResolver is simple decorator for Resolver
// to cache result in LRU.
type LRUResolver struct {
	next Resolver

	capacity int
	cache    map[string]*linkedNode
	lruList  *linkedList
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

// Resolve implements Resolver.
func (l *LRUResolver) Resolve(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	// TODO(tdakkota): expiration support
	if v, ok := l.get(domain); ok {
		return v, nil
	}

	r, err := l.next.Resolve(ctx, domain)
	if err != nil {
		return nil, err
	}

	l.put(domain, r)
	return r, nil
}

func (l *LRUResolver) get(key string) (v tg.InputPeerClass, ok bool) {
	if found, ok := l.cache[key]; ok {
		l.lruList.MoveToFront(found)
		return found.value, true
	}
	return
}

func (l *LRUResolver) put(key string, value tg.InputPeerClass) {
	if l.capacity == 0 {
		return
	}

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
