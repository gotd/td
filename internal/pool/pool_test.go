package pool

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/tdsync"
)

type invokerFunc func(ctx context.Context, input bin.Encoder, output bin.Decoder) error

type mockConn struct {
	ready  *tdsync.Ready
	stop   *tdsync.Ready
	locker *sync.RWMutex
	invoke invokerFunc
}

func (mockConn) Ping(ctx context.Context) error {
	return errors.New("not implemented")
}

func newMockConn(invoke invokerFunc) mockConn {
	return mockConn{
		ready:  tdsync.NewReady(),
		stop:   tdsync.NewReady(),
		locker: new(sync.RWMutex),
		invoke: invoke,
	}
}

func (m mockConn) Run(ctx context.Context) error {
	m.ready.Signal()
	select {
	case <-m.stop.Ready():
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func (m mockConn) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	m.locker.RLock()
	defer m.locker.RUnlock()
	return m.invoke(ctx, input, output)
}

func (m mockConn) Ready() <-chan struct{} {
	return m.ready.Ready()
}

func (m mockConn) lock() sync.Locker {
	m.locker.Lock()
	return m.locker
}

func (m mockConn) kill() {
	m.stop.Signal()
}

type connBuilder struct {
	conns   []mockConn
	lockers map[int]sync.Locker
	mux     sync.Mutex

	invoke invokerFunc
	ctx    context.Context
}

func newConnBuilder(ctx context.Context, invoke invokerFunc) *connBuilder {
	return &connBuilder{invoke: invoke, ctx: ctx, lockers: map[int]sync.Locker{}}
}

func (c *connBuilder) create() Conn {
	c.mux.Lock()
	defer c.mux.Unlock()

	i := len(c.conns)
	c.conns = append(c.conns, newMockConn(c.invoke))
	return c.conns[i]
}

func (c *connBuilder) lockOne() {
	c.mux.Lock()

	if len(c.conns) == 0 {
		c.mux.Unlock()
	loop:
		for {
			select {
			case <-c.ctx.Done():
				panic(c.ctx.Err())
			default:
				c.mux.Lock()
				if len(c.conns) != 0 {
					break loop
				}
				c.mux.Unlock()

				runtime.Gosched()
			}
		}
	}

	defer c.mux.Unlock()
	var n int
	for {
		n = rand.Intn(len(c.conns))
		_, ok := c.lockers[n]
		if !ok {
			break
		}
	}

	locker := c.conns[n].lock()
	c.lockers[n] = locker
}

func (c *connBuilder) unlockOne() {
	c.mux.Lock()
	defer c.mux.Unlock()

	var idx int
	var locker sync.Locker
	for idx, locker = range c.lockers {
		break
	}

	if locker == nil {
		panic("no lockers")
	}

	delete(c.lockers, idx)
	locker.Unlock()
}

func (c *connBuilder) killOne() {
	c.mux.Lock()
	defer c.mux.Unlock()

	if len(c.conns) == 0 {
		return
	}

	n := rand.Intn(len(c.conns))
	c.conns[n].kill()

	// Delete from SliceTricks.
	copy(c.conns[n:], c.conns[n+1:])
	c.conns[len(c.conns)-1] = mockConn{}
	c.conns = c.conns[:len(c.conns)-1]
}

type script []byte

func runScript(open int64, s script) func(t *testing.T) {
	return func(t *testing.T) {
		a := require.New(t)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log := zaptest.NewLogger(t)

		b := newConnBuilder(ctx, func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
			return nil
		})
		dc := NewDC(ctx, 2, b.create, DCOptions{
			Logger:             log.Named("dc"),
			MaxOpenConnections: open,
		})
		defer dc.Close()

		wg := tdsync.NewCancellableGroup(ctx)
		for _, action := range s {
			switch action {
			case 'i':
				a.NoError(dc.Invoke(ctx, nil, nil))
			case 'a':
				wg.Go(func(ctx context.Context) error {
					return dc.Invoke(ctx, nil, nil)
				})
			case 'k':
				b.killOne()
			case 'l':
				b.lockOne()
			case 'u':
				b.unlockOne()
			default:
				t.Fatalf("Invalid action %c", action)
			}
		}
		a.NoError(wg.Wait())
	}
}

func testAllScenario(open int64) func(t *testing.T) {
	return func(t *testing.T) {
		scripts := []struct {
			name    string
			code    string
			minConn int64
		}{
			{"", "iki", 0},
			{"", "ikii", 0},
			{"", "ilaaui", 2},
			{"", "alaaui", 2},
			{"", "ilkaaui", 2},
			{"", "ilkaui", 2},
		}

		for _, sc := range scripts {
			if open != 0 && open < sc.minConn {
				continue
			}

			var s strings.Builder
			if sc.name == "" {
				for _, action := range []byte(sc.code) {
					switch action {
					case 'i':
						s.WriteString("Call")
					case 'a':
						s.WriteString("Async")
					case 'k':
						s.WriteString("Kill")
					case 'l':
						s.WriteString("Lock")
					case 'u':
						s.WriteString("Unlock")
					default:
						t.Fatalf("Invalid action %c", action)
					}
				}
			}
			t.Run(s.String(), runScript(open, []byte(sc.code)))
		}
	}
}

func TestDC(t *testing.T) {
	limits := []int64{0, 1, 2, 4}
	for _, limit := range limits {
		t.Run(fmt.Sprintf("Conns%d", limit), testAllScenario(limit))
	}
}
