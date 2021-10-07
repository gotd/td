package rpc

import (
	"bytes"
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/nnqq/td/bin"
)

// mockObject implements bin.Object for testing.
type mockObject struct {
	data []byte
}

func (m mockObject) Decode(b *bin.Buffer) error {
	if !bytes.Equal(b.Buf, m.data) {
		return errors.New("mismatch")
	}
	b.Skip(len(b.Buf))
	return nil
}

func (m mockObject) Encode(b *bin.Buffer) error {
	b.Put(m.data)
	return nil
}

func BenchmarkEngine_Do(b *testing.B) {
	ids := make(chan int64, 100)
	defer close(ids)

	e := New(func(ctx context.Context, msgID int64, seqNo int32, in bin.Encoder) error {
		ids <- msgID
		return nil
	}, Options{})

	var id int64

	ctx := context.Background()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		go func() {
			buf := &bin.Buffer{}
			// Fake handler.
			obj := mockObject{data: make([]byte, 100)}

			for id := range ids {
				e.NotifyAcks([]int64{id})

				buf.ResetTo(obj.data)
				if err := e.NotifyResult(id, buf); err != nil {
					b.Error(err)
				}
			}
		}()

		obj := mockObject{data: make([]byte, 100)}

		for pb.Next() {
			nextID := atomic.AddInt64(&id, 1)
			if err := e.Do(ctx, Request{
				MsgID:  nextID,
				Input:  obj,
				Output: obj,
			}); err != nil {
				b.Fatal(err)
			}
		}
	})
}
