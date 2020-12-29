package telegram

import (
	"sync"
	"time"

	"github.com/gotd/td/internal/proto"
)

// msgIDGen is message id generator that provides collision prevention.
type msgIDGen struct {
	t      proto.MessageType
	buf    []int64
	cursor int
	mux    sync.Mutex
	now    func() time.Time
}

func (g *msgIDGen) contains(id int64) bool {
	for i := range g.buf {
		if g.buf[i] == id {
			return true
		}
	}
	return false
}

func (g *msgIDGen) add(id int64) {
	if g.cursor >= len(g.buf) {
		g.cursor = 0
	}
	g.buf[g.cursor] = id
	g.cursor++
}

func (g *msgIDGen) New() int64 {
	g.mux.Lock()
	defer g.mux.Unlock()

	now := g.now()
	id := int64(proto.NewMessageID(now, g.t))

	const resolution = time.Nanosecond * 5

	for g.contains(id) {
		now = now.Add(resolution)
		id = int64(proto.NewMessageID(now, g.t))
	}

	g.add(id)

	return id
}

func newMsgIDGen(now func() time.Time, n int, t proto.MessageType) *msgIDGen {
	return &msgIDGen{
		buf: make([]int64, n),
		now: now,
		t:   t,
	}
}

func (c *Client) newMessageID() int64 {
	return c.msgID.New()
}
