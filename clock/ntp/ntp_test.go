package ntp

import (
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/clock"
)

func TestNew(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		a := require.New(t)
		c, err := New(Options{
			query: func(server string) (time.Duration, error) {
				a.Equal(DefaultServer, server)
				return time.Second, nil
			},
		})
		a.NoError(err)
		a.Equal(time.Second, c.Offset())
	})

	t.Run("SyncError", func(t *testing.T) {
		a := require.New(t)
		_, err := New(Options{
			query: func(string) (time.Duration, error) {
				return 0, errors.New("boom")
			},
		})
		a.Error(err)
	})
}

func TestClock_Now(t *testing.T) {
	a := require.New(t)
	base := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	c, err := New(Options{
		Server: "example.com:123",
		Now:    func() time.Time { return base },
		query: func(server string) (time.Duration, error) {
			a.Equal("example.com:123", server)
			return time.Minute, nil
		},
	})
	a.NoError(err)

	a.Equal(time.Minute, c.Offset())
	a.Equal(base.Add(time.Minute), c.Now())
}

func TestClock_Sync(t *testing.T) {
	a := require.New(t)
	offset := time.Second
	fail := false
	c, err := New(Options{
		query: func(string) (time.Duration, error) {
			if fail {
				return 0, errors.New("boom")
			}
			return offset, nil
		},
	})
	a.NoError(err)
	a.Equal(time.Second, c.Offset())

	// Re-sync with a new offset.
	offset = 2 * time.Second
	a.NoError(c.Sync())
	a.Equal(2*time.Second, c.Offset())

	// On error the previous offset is kept.
	fail = true
	a.Error(c.Sync())
	a.Equal(2*time.Second, c.Offset())
}

func TestClock_Interface(t *testing.T) {
	a := require.New(t)
	c, err := New(Options{
		query: func(string) (time.Duration, error) { return 0, nil },
	})
	a.NoError(err)

	var cl clock.Clock = c
	timer := cl.Timer(time.Hour)
	a.NotNil(timer)
	clock.StopTimer(timer)

	ticker := cl.Ticker(time.Hour)
	a.NotNil(ticker)
	ticker.Stop()
}
