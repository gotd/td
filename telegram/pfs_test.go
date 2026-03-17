package telegram

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/exchange"
	"github.com/gotd/td/mtproto"
	"github.com/gotd/td/pool"
)

func TestClientHandleDCConnDeadPFSDropResetsSession(t *testing.T) {
	a := require.New(t)
	client := Client{
		log: zap.NewNop(),
	}
	client.init()

	dcID := 5
	key := crypto.Key{1}.WithID()
	session := pool.NewSyncSession(pool.Session{
		DC:      dcID,
		AuthKey: key,
		Salt:    77,
	})
	client.sessions[dcID] = session

	onDeadCalls := 0
	client.onDead = func(error) {
		onDeadCalls++
	}

	client.handleDCConnDead(dcID, mtproto.ErrPFSDropKeysRequired)

	// PFS drop request forces cached key reset.
	data := session.Load()
	a.True(data.AuthKey.Zero())
	a.Zero(data.Salt)
	a.Equal(1, onDeadCalls)
}

func TestClientHandleDCConnDeadPassThrough(t *testing.T) {
	a := require.New(t)
	client := Client{
		log: zap.NewNop(),
	}
	client.init()

	dcID := 6
	key := crypto.Key{2}.WithID()
	session := pool.NewSyncSession(pool.Session{
		DC:      dcID,
		AuthKey: key,
		Salt:    88,
	})
	client.sessions[dcID] = session

	testErr := errors.New("test")
	onDeadCalls := 0
	client.onDead = func(err error) {
		a.Equal(testErr, err)
		onDeadCalls++
	}

	client.handleDCConnDead(dcID, testErr)

	// Non-PFS error must not mutate stored auth key.
	data := session.Load()
	a.Equal(key, data.AuthKey)
	a.Equal(int64(88), data.Salt)
	a.Equal(1, onDeadCalls)
}

func TestClientHandleCDNConnDeadPFSDropResetsCDNSession(t *testing.T) {
	a := require.New(t)
	client := Client{
		log: zap.NewNop(),
	}
	client.init()

	dcID := 7
	key := crypto.Key{3}.WithID()
	session := pool.NewSyncSession(pool.Session{
		DC:      dcID,
		AuthKey: key,
		Salt:    99,
	})
	client.cdnSessions[dcID] = session

	onDeadCalls := 0
	client.onDead = func(error) {
		onDeadCalls++
	}

	client.handleCDNConnDead(dcID, mtproto.ErrPFSDropKeysRequired)

	data := session.Load()
	a.True(data.AuthKey.Zero())
	a.Zero(data.Salt)
	a.Equal(1, onDeadCalls)
}

func TestClientHandleCDNConnDeadDoesNotTouchRegularSession(t *testing.T) {
	a := require.New(t)
	client := Client{
		log: zap.NewNop(),
	}
	client.init()

	dcID := 8
	regularKey := crypto.Key{4}.WithID()
	cdnKey := crypto.Key{5}.WithID()
	client.sessions[dcID] = pool.NewSyncSession(pool.Session{
		DC:      dcID,
		AuthKey: regularKey,
		Salt:    11,
	})
	client.cdnSessions[dcID] = pool.NewSyncSession(pool.Session{
		DC:      dcID,
		AuthKey: cdnKey,
		Salt:    22,
	})

	client.handleCDNConnDead(dcID, mtproto.ErrPFSDropKeysRequired)

	regular := client.sessions[dcID].Load()
	cdn := client.cdnSessions[dcID].Load()
	a.Equal(regularKey, regular.AuthKey)
	a.Equal(int64(11), regular.Salt)
	a.True(cdn.AuthKey.Zero())
	a.Zero(cdn.Salt)
}

type closeInvokerStub struct {
	closed   bool
	closedCh chan struct{}
}

func (*closeInvokerStub) Invoke(context.Context, bin.Encoder, bin.Decoder) error {
	return nil
}

func (s *closeInvokerStub) Close() error {
	s.closed = true
	if s.closedCh != nil {
		close(s.closedCh)
	}
	return nil
}

type blockingCloseInvokerStub struct {
	closed chan struct{}
	unlock chan struct{}
}

func (*blockingCloseInvokerStub) Invoke(context.Context, bin.Encoder, bin.Decoder) error {
	return nil
}

func (s *blockingCloseInvokerStub) Close() error {
	close(s.closed)
	<-s.unlock
	return nil
}

func TestClientHandleCDNConnDeadFingerprintMissInvalidatesCache(t *testing.T) {
	a := require.New(t)
	client := Client{
		log: zap.NewNop(),
	}
	client.init()

	const dcID = 9
	conn := &closeInvokerStub{closedCh: make(chan struct{})}
	client.cdnPools.conns[dcID] = []cachedCDNPool{{
		conn: conn,
		max:  1,
	}}
	client.cdnKeysSet = true
	client.cdnKeys = []PublicKey{{}}

	onDeadCalls := 0
	client.onDead = func(error) {
		onDeadCalls++
	}

	client.handleCDNConnDead(dcID, exchange.ErrKeyFingerprintNotFound)

	client.cdnPools.mux.Lock()
	_, ok := client.cdnPools.conns[dcID]
	client.cdnPools.mux.Unlock()
	a.False(ok)
	select {
	case <-conn.closedCh:
	case <-time.After(time.Second):
		t.Fatal("expected async close call")
	}
	a.True(conn.closed)
	a.False(client.cdnKeysSet)
	a.Nil(client.cdnKeys)
	a.Equal(0, onDeadCalls, "fingerprint miss should be handled internally without onDead callback")
}

func TestClientHandleCDNConnDeadFingerprintMissDoesNotBlockOnClose(t *testing.T) {
	client := Client{
		log: zap.NewNop(),
	}
	client.init()

	const dcID = 10
	conn := &blockingCloseInvokerStub{
		closed: make(chan struct{}),
		unlock: make(chan struct{}),
	}
	client.cdnPools.conns[dcID] = []cachedCDNPool{{
		conn: conn,
		max:  1,
	}}

	done := make(chan struct{})
	go func() {
		client.handleCDNConnDead(dcID, exchange.ErrKeyFingerprintNotFound)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handleCDNConnDead blocked on pool close")
	}

	select {
	case <-conn.closed:
	case <-time.After(time.Second):
		t.Fatal("expected async close call")
	}
	close(conn.unlock)
}

type observedBlockingCloseInvokerStub struct {
	started chan struct{}
	unlock  chan struct{}
}

func (*observedBlockingCloseInvokerStub) Invoke(context.Context, bin.Encoder, bin.Decoder) error {
	return nil
}

func (s *observedBlockingCloseInvokerStub) Close() error {
	select {
	case s.started <- struct{}{}:
	default:
	}
	<-s.unlock
	return nil
}

func TestClientHandleCDNConnDeadFingerprintMissStartsMultipleCloseWorkers(t *testing.T) {
	client := Client{
		log: zap.NewNop(),
	}
	client.init()

	const (
		dcID         = 11
		totalStale   = 32
		maxWorkers   = 4
		minWorkers   = 2
		waitForStart = time.Second
	)
	started := make(chan struct{}, totalStale)
	unlock := make(chan struct{})
	pools := make([]cachedCDNPool, 0, totalStale)
	for i := 0; i < totalStale; i++ {
		pools = append(pools, cachedCDNPool{
			conn: &observedBlockingCloseInvokerStub{
				started: started,
				unlock:  unlock,
			},
			max: int64(i + 1),
		})
	}
	client.cdnPools.conns[dcID] = pools

	client.handleCDNConnDead(dcID, exchange.ErrKeyFingerprintNotFound)

	startedCount := 0
	deadline := time.After(waitForStart)
	for startedCount < minWorkers {
		select {
		case <-started:
			startedCount++
		case <-deadline:
			t.Fatalf("expected at least %d close workers to start, got %d", minWorkers, startedCount)
		}
	}

	// Workers are blocked in Close(); after a short wait the amount of started
	// workers should stay bounded.
	time.Sleep(100 * time.Millisecond)
	select {
	case <-started:
		startedCount++
	default:
	}
	for {
		select {
		case <-started:
			startedCount++
		default:
			goto done
		}
	}
done:
	if startedCount > maxWorkers {
		t.Fatalf("expected bounded close workers <= %d, got %d", maxWorkers, startedCount)
	}

	close(unlock)
}

func TestClientHandleCDNConnDeadFingerprintMissProcessesMultipleCallsInParallel(t *testing.T) {
	client := Client{
		log: zap.NewNop(),
	}
	client.init()

	started := make(chan struct{}, 2)
	unlock := make(chan struct{})
	first := &observedBlockingCloseInvokerStub{started: started, unlock: unlock}
	second := &observedBlockingCloseInvokerStub{started: started, unlock: unlock}

	client.cdnPools.conns[12] = []cachedCDNPool{{conn: first, max: 1}}
	client.handleCDNConnDead(12, exchange.ErrKeyFingerprintNotFound)

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("expected first close call")
	}

	client.cdnPools.conns[13] = []cachedCDNPool{{conn: second, max: 1}}
	client.handleCDNConnDead(13, exchange.ErrKeyFingerprintNotFound)

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("expected second close call to start without waiting first close")
	}

	close(unlock)
}
