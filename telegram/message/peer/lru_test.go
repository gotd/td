package peer

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/gotd/neo"

	"github.com/nnqq/td/internal/testutil"
	"github.com/nnqq/td/tg"
)

func TestLRU(t *testing.T) {
	a := require.New(t)
	c := neo.NewTime(time.Now())
	lru := NewLRUResolver(nil, 4).WithClock(c).WithExpiration(1 * time.Second)

	// Add 5 entries.
	// State: [4 3 2 1]
	for i := range [5]struct{}{} {
		lru.put(strconv.Itoa(i), &tg.InputPeerChat{
			ChatID: int64(i),
		})
	}

	// First entry must be evicted.
	// State: [4 3 2 1]
	_, ok := lru.get(strconv.Itoa(0))
	a.False(ok)

	// Third must not.
	// State: [2 3 4 1]
	_, ok = lru.get(strconv.Itoa(2))
	a.True(ok)
	a.Equal("2", lru.lruList.Front().nodeData.key)

	// Add yet another.
	// State: [6 2 3 4]
	lru.put(strconv.Itoa(6), &tg.InputPeerChat{
		ChatID: 6,
	})
	a.Equal("6", lru.lruList.Front().nodeData.key)

	// Then yet one must be evicted.
	_, ok = lru.get(strconv.Itoa(1))
	a.False(ok)

	// Add which already exist.
	// State: [4 6 2 3]
	lru.put(strconv.Itoa(4), &tg.InputPeerChat{
		ChatID: 6,
	})
	a.Equal("4", lru.lruList.Front().nodeData.key)

	// Delete key which does not exist.
	// State: [4 6 2 3]
	_, ok = lru.Evict(strconv.Itoa(10))
	a.False(ok)

	c.Travel(time.Hour)
	// Delete expired key.
	// State: [6 2 3]
	_, ok = lru.get("4")
	a.False(ok)
}

type mockResolver struct {
	counter       int
	returnErr     bool
	domain, phone string
	peer          tg.InputPeerClass
	t             testing.TB
}

func (m *mockResolver) ResolveDomain(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	m.counter++

	if m.returnErr && m.counter == 1 {
		return nil, testutil.TestError()
	}

	if domain != m.domain {
		err := xerrors.Errorf("expected domain %q, got %q", m.domain, domain)
		m.t.Error(err)
		return nil, err
	}
	return m.peer, nil
}

func (m *mockResolver) ResolvePhone(ctx context.Context, phone string) (tg.InputPeerClass, error) {
	m.counter++

	if m.returnErr && m.counter == 1 {
		return nil, testutil.TestError()
	}

	if phone != m.phone {
		err := xerrors.Errorf("expected phone %q, got %q", m.phone, phone)
		m.t.Error(err)
		return nil, err
	}
	return m.peer, nil
}

func TestLRUResolver_Resolve(t *testing.T) {
	ctx := context.Background()
	expectedDomain := "telegram"
	expected := &tg.InputPeerUser{
		UserID: 10,
	}

	t.Run("Cache", func(t *testing.T) {
		a := require.New(t)
		resolver := &mockResolver{
			domain: expectedDomain,
			peer:   expected,
			t:      t,
		}

		lru := NewLRUResolver(resolver, 10)

		r, err := lru.ResolveDomain(ctx, expectedDomain)
		a.NoError(err)
		a.Equal(expected, r)

		r2, err := lru.ResolveDomain(ctx, expectedDomain)
		a.NoError(err)
		a.Equal(expected, r2)

		a.Equalf(1, resolver.counter, "RPC call was not cached")
	})

	t.Run("Error", func(t *testing.T) {
		a := require.New(t)
		resolver := &mockResolver{
			returnErr: true,
			domain:    expectedDomain,
			peer:      expected,
			t:         t,
		}

		lru := NewLRUResolver(resolver, 10)

		_, err := lru.ResolveDomain(ctx, expectedDomain)
		a.Error(err)

		r2, err := lru.ResolveDomain(ctx, expectedDomain)
		a.NoError(err)
		a.Equal(expected, r2)

		a.Equalf(2, resolver.counter, "RPC call error was cached")
	})
}
