package mtproto

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/tg"
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

	const (
		workers           = 32
		requestsPerWorker = 200
	)
	client := newTestClient(func(msgID int64, seqNo int32, body bin.Encoder) (bin.Encoder, error) {
		mux.Lock()
		records = append(records, record{msgID, seqNo})
		mux.Unlock()
		return &tg.Config{}, nil
	})

	{
		var wg sync.WaitGroup
		wg.Add(workers)
		for i := 0; i < workers; i++ {
			go func(t *testing.T) {
				defer wg.Done()

				for i := 0; i < requestsPerWorker; i++ {
					err := client.Invoke(context.Background(), &tg.Config{}, &tg.Config{})
					require.NoError(t, err)
				}
			}(t)
		}

		wg.Wait()
	}

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

	for _, current := range records {
		if contains(current.msgID) {
			// Ignore duplicates.
			continue
		}

		less(current.msgID, func(less record) {
			// The server has already received a message with a lower msg_id
			// but with either a higher or an equal and odd seqno.
			if less.seqNo >= current.seqNo && less.seqNo%2 == 1 {
				t.Fatal("seqNo too low")
			}
		})

		greater(current.msgID, func(greater record) {
			// Similarly, there is a message with a higher msg_id
			// but with either a lower or an equal and odd seqno.
			if greater.seqNo <= current.seqNo && greater.seqNo%2 == 1 {
				t.Fatal("seqNo too high")
			}
		})

		received = append(received, current)
	}
}
