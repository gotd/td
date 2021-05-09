package floodwait

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/neo"
)

func TestScheduler(t *testing.T) {
	a := require.New(t)
	n := neo.NewTime(time.Now())
	sch := newScheduler(n, time.Second)

	r := request{
		key: 1,
	}
	// Schedule request.
	sch.schedule(r)
	// Got flood wait.
	sch.flood(r, 5*time.Second)
	// Ensure that request re-scheduled.
	a.Empty(sch.gather(nil))

	// Travel and ensure that request re-scheduled correctly.
	n.Travel(5*time.Second + time.Millisecond)
	a.Len(sch.gather(nil), 1)

	// Decrease wait timeout.
	sch.nice(1)
	// Schedule yet one
	sch.schedule(request{
		key: 1,
	})
	// Ensure that timer decreased correctly.
	n.Travel(4*time.Second + time.Millisecond)
	a.Len(sch.gather(nil), 1)
}
