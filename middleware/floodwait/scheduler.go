package floodwait

import (
	"context"
	"sync"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
)

type scheduler struct {
	state map[key]time.Duration
	mux   sync.Mutex
	queue *queue

	clock clock.Clock
	dec   time.Duration
}

func newScheduler(c clock.Clock, dec time.Duration) *scheduler {
	const initialCapacity = 16

	return &scheduler{
		state: make(map[key]time.Duration, initialCapacity),
		queue: newQueue(initialCapacity),
		clock: c,
		dec:   dec,
	}
}

func (s *scheduler) new(ctx context.Context, input bin.Encoder, output bin.Decoder) <-chan error {
	var k key
	k.fromEncoder(input)
	r := request{
		ctx:    ctx,
		input:  input,
		output: output,
		key:    k,
		result: make(chan error, 1),
	}

	s.mux.Lock()
	defer s.mux.Unlock()
	s.schedule(r)
	return r.result
}

// schedule adds request to the queue.
// Assumes the mutex is locked.
func (s *scheduler) schedule(r request) {
	k := r.key

	var t time.Time
	if state, ok := s.state[k]; ok {
		t = s.clock.Now().Add(state)
	} else {
		t = s.clock.Now()
	}
	s.queue.add(r, t)
}

func (s *scheduler) gather(r []scheduled) []scheduled {
	return s.queue.gather(s.clock.Now(), r)
}

func (s *scheduler) nice(k key) {
	s.mux.Lock()
	if state, ok := s.state[k]; ok && state-s.dec > 0 {
		s.state[k] = state - s.dec
	} else {
		delete(s.state, k)
	}
	s.mux.Unlock()
}

func (s *scheduler) flood(req request, d time.Duration) {
	k := req.key

	s.mux.Lock()
	now := s.clock.Now()
	if state, ok := s.state[k]; !ok || state < d {
		s.state[k] = d
	}
	s.queue.add(req, now.Add(d))
	s.mux.Unlock()

	s.queue.move(k, now, d)
}
