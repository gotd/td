package mtproto

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestMsgSeq(t *testing.T) {
	type record struct {
		msgID int64
		seqNo int32
	}

	var (
		records []record
		mux     sync.Mutex
	)

	conn := newTestClient(func(msgID int64, seqNo int32, body bin.Encoder) (bin.Encoder, error) {
		mux.Lock()
		records = append(records, record{msgID, seqNo})
		mux.Unlock()
		return &tg.Config{}, nil
	})

	const (
		workers           = 32
		requestsPerWorker = 200
	)

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(t *testing.T) {
			defer wg.Done()

			for i := 0; i < requestsPerWorker; i++ {
				err := conn.InvokeRaw(context.Background(), &tg.Config{}, &tg.Config{})
				require.NoError(t, err)
			}
		}(t)
	}

	wg.Wait()

	var received []record
	less := func(msgID int64, f func(r record)) {
		for _, recv := range received {
			if recv.msgID < msgID {
				f(recv)
			}
		}
	}

	greater := func(msgID int64, f func(r record)) {
		for _, recv := range received {
			if recv.msgID > msgID {
				f(recv)
			}
		}
	}

	contains := func(msgID int64) bool {
		for _, rec := range received {
			if rec.msgID == msgID {
				return true
			}
		}

		return false
	}

	for _, recv := range records {
		if contains(recv.msgID) {
			continue
		}

		less(recv.msgID, func(r record) {
			seq := r.seqNo
			if seq >= recv.seqNo && seq%2 == 1 {
				t.Fatal("seqNo too low")
			}
		})

		greater(recv.msgID, func(r record) {
			seq := r.seqNo
			if seq <= recv.seqNo && seq%2 == 1 {
				t.Fatal("seqNo too high")
			}
		})

		received = append(received, recv)
	}
}
