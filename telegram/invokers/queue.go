package invokers

import (
	"container/heap"
	"sync"
	"time"
)

type scheduled struct {
	request

	sendTime time.Time
	index    int
}

type scheduledHeap []scheduled

func (r scheduledHeap) Len() int { return len(r) }

func (r scheduledHeap) Less(i, j int) bool {
	return r[i].sendTime.Before(r[j].sendTime)
}

func (r scheduledHeap) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
	r[i].index = i
	r[j].index = j
}

func (r *scheduledHeap) Push(x interface{}) {
	n := len(*r)
	item := x.(scheduled)
	item.index = n
	*r = append(*r, item)
}

func (r *scheduledHeap) Pop() interface{} {
	old := *r
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*r = old[0 : n-1]
	return item
}

func (r *scheduledHeap) Peek() (scheduled, bool) {
	old := *r
	n := len(old)
	if n < 1 {
		return scheduled{}, false
	}
	x := old[n-1]
	return x, true
}

type queue struct {
	requests    *scheduledHeap
	requestsMux sync.Mutex
}

func newQueue(initialCapacity int) *queue {
	r := make(scheduledHeap, 0, initialCapacity)
	return &queue{requests: &r}
}

func (q *queue) add(r request, t time.Time) {
	q.requestsMux.Lock()
	defer q.requestsMux.Unlock()
	heap.Push(q.requests, scheduled{
		request:  r,
		sendTime: t,
	})
}

func (q *queue) move(k key, dur time.Duration) {
	q.requestsMux.Lock()
	defer q.requestsMux.Unlock()

	old := *q.requests
	for idx := range old {
		if old[idx].key != k {
			continue
		}

		t := old[idx].sendTime
		old[idx].sendTime = t.Add(dur)
		heap.Fix(q.requests, idx)
	}
}

func (q *queue) gather(now time.Time, req []scheduled) []scheduled {
	q.requestsMux.Lock()
	defer q.requestsMux.Unlock()

	for {
		next, ok := q.requests.Peek()
		if !ok {
			return req
		}

		switch {
		case !ok:
			return req
		case now.Before(next.sendTime):
			return req
		default:
			heap.Pop(q.requests)
			req = append(req, next)
		}
	}
}
