package invokers

import (
	"context"
	"sync"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
)

type key uint64

func (k *key) fromEncoder(encoder bin.Encoder) {
	obj, ok := encoder.(Object)
	if !ok {
		return
	}
	*k = key(obj.TypeID())
}

type request struct {
	ctx    context.Context
	input  bin.Encoder
	output bin.Decoder
	key    key

	result chan error
}

type scheduler struct {
	state map[key]time.Duration
	mux   sync.Mutex
	queue *queue

	clock clock.Clock
	dec   time.Duration
}

func newScheduler(c clock.Clock, dec time.Duration) *scheduler {
	return &scheduler{
		state: map[key]time.Duration{},
		queue: newQueue(16),
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
	}
	s.mux.Unlock()
}

func (s *scheduler) flood(req request, d time.Duration) {
	k := req.key

	s.mux.Lock()
	if state, ok := s.state[k]; !ok || state < d {
		s.state[k] = d
	}
	s.queue.add(req, s.clock.Now().Add(d))
	s.mux.Unlock()

	s.queue.move(k, d)
}
