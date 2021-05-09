package floodwait

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	a := require.New(t)

	q := newQueue(16)
	now := time.Date(2077, 10, 23, 0, 3, 0, 0, time.UTC)
	for i := range [10]struct{}{} {
		q.add(request{
			key: key(i),
		}, now.Add(time.Duration(i)*time.Second))
	}

	a.Equal(10, q.len())

	a.Len(q.gather(now.Add(1*time.Millisecond), nil), 1)
	a.Equal(9, q.len())

	now = now.Add(10 * time.Second)
	q.move(5, now, 10*time.Second)

	a.Len(q.gather(now, nil), 8)
	a.Equal(1, q.len())

	a.Len(q.gather(now.Add(10*time.Second), nil), 1)
	a.Equal(0, q.len())
}
