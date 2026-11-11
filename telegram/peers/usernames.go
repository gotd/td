package peers

import (
	"strings"
	"sync"
	"time"

	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
)

// UsernameTTL is how long Manager trusts a cached username: TDLib keeps
// resolved usernames for a day (USERNAME_CACHE_EXPIRE_TIME).
const UsernameTTL = 24 * time.Hour

// NormalizeUsername returns the key a username is cached under: without a
// leading "@" and dots, lowercased, like TDLib's clean_username. Telegram
// usernames are case-insensitive.
func NormalizeUsername(s string) string {
	s = strings.TrimPrefix(strings.TrimSpace(s), "@")
	s = strings.ReplaceAll(s, ".", "")
	return strings.ToLower(s)
}

type usernameEntry struct {
	peer  tg.PeerClass
	until time.Time
}

// usernameCache maps normalized usernames to peers.
type usernameCache struct {
	mux    sync.Mutex
	names  map[string]usernameEntry
	byPeer map[constant.TDLibPeerID][]string
}

func newUsernameCache() *usernameCache {
	return &usernameCache{
		names:  map[string]usernameEntry{},
		byPeer: map[constant.TDLibPeerID][]string{},
	}
}

// get returns the peer of a normalized username if it has not expired.
func (c *usernameCache) get(name string, now time.Time) (tg.PeerClass, bool) {
	c.mux.Lock()
	defer c.mux.Unlock()
	e, ok := c.names[name]
	if !ok || !now.Before(e.until) {
		return nil, false
	}
	return e.peer, true
}

// set replaces the usernames of a peer: names the peer no longer has are
// dropped, so a renamed peer is not found by its old name.
func (c *usernameCache) set(id constant.TDLibPeerID, peer tg.PeerClass, names []string, until time.Time) {
	c.mux.Lock()
	defer c.mux.Unlock()
	keep := make(map[string]struct{}, len(names))
	for _, n := range names {
		if n = NormalizeUsername(n); n != "" {
			keep[n] = struct{}{}
		}
	}
	for _, old := range c.byPeer[id] {
		if _, ok := keep[old]; ok {
			continue
		}
		if e, ok := c.names[old]; ok && peerID(e.peer) == id {
			delete(c.names, old)
		}
	}
	list := make([]string, 0, len(keep))
	for n := range keep {
		c.names[n] = usernameEntry{peer: peer, until: until}
		list = append(list, n)
	}
	if len(list) == 0 {
		delete(c.byPeer, id)
		return
	}
	c.byPeer[id] = list
}

// add adds one more username of a peer, e.g. a name it was resolved by.
func (c *usernameCache) add(id constant.TDLibPeerID, peer tg.PeerClass, name string, until time.Time) {
	name = NormalizeUsername(name)
	if name == "" {
		return
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	c.names[name] = usernameEntry{peer: peer, until: until}
	for _, n := range c.byPeer[id] {
		if n == name {
			return
		}
	}
	c.byPeer[id] = append(c.byPeer[id], name)
}

// drop forgets a username, e.g. on USERNAME_NOT_OCCUPIED.
func (c *usernameCache) drop(name string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	delete(c.names, name)
}

func peerID(p tg.PeerClass) constant.TDLibPeerID {
	switch p := p.(type) {
	case *tg.PeerUser:
		return userPeerID(p.UserID)
	case *tg.PeerChat:
		return chatPeerID(p.ChatID)
	case *tg.PeerChannel:
		return channelPeerID(p.ChannelID)
	}
	return 0
}

// activeUsernames returns the main username and the active ones of
// usernames[] (collectible and additional usernames).
func activeUsernames(main string, list []tg.Username) []string {
	out := make([]string, 0, 1+len(list))
	if main != "" {
		out = append(out, main)
	}
	for _, u := range list {
		if u.Active {
			out = append(out, u.Username)
		}
	}
	return out
}
