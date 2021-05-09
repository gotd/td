package floodwait

import (
	"container/heap"
	"sync"
	"time"
)

type scheduled struct {
	request  request
	sendTime time.Time
	index    int
}

type scheduledHeap []scheduled

func (r scheduledHeap) Len() int { return len(r) }

func (r scheduledHeap) Less(i, j int) bool {
	return r[i].sendTime.UnixNano() < r[j].sendTime.UnixNano()
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

type queue struct {
	requests    scheduledHeap
	requestsMux sync.Mutex
}

func newQueue(initialCapacity int) *queue {
	r := make(scheduledHeap, 0, initialCapacity)
	return &queue{requests: r}
}

func (q *queue) add(r request, t time.Time) {
	q.requestsMux.Lock()
	defer q.requestsMux.Unlock()

	heap.Push(&q.requests, scheduled{
		request:  r,
		sendTime: t,
	})
}

func (q *queue) len() int {
	q.requestsMux.Lock()
	r := len(q.requests)
	q.requestsMux.Unlock()
	return r
}

func (q *queue) move(k key, now time.Time, dur time.Duration) {
	q.requestsMux.Lock()
	defer q.requestsMux.Unlock()

	for idx, s := range q.requests {
		if s.request.key != k {
			continue
		}

		t := s.sendTime
		if t.Sub(now) > dur {
			break
		}
		q.requests[idx].sendTime = t.Add(dur)
	}
	heap.Init(&q.requests)
}

func (q *queue) gather(now time.Time, req []scheduled) []scheduled {
	q.requestsMux.Lock()
	defer q.requestsMux.Unlock()

	for {
		if q.requests.Len() < 1 {
			return req
		}

		next := heap.Pop(&q.requests).(scheduled)
		if now.Before(next.sendTime) {
			heap.Push(&q.requests, next)
			return req
		}
		req = append(req, next)
	}
}
