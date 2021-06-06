package updates

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestSequenceBox(t *testing.T) {
	var (
		state   int
		updates []update
	)

	box := newSequenceBox(sequenceConfig{
		InitialState: 3,
		Apply: func(s int, u []update) error {
			state = s
			updates = append(updates, u...)
			return nil
		},
		OnGap:  func() {},
		Logger: zaptest.NewLogger(t),
	})

	require.Nil(t, box.Handle(update{
		Value: 1,
		State: 2,
		Count: 1,
	}))
	require.Zero(t, state)
	require.Empty(t, updates)
	require.Empty(t, box.pending)

	require.Nil(t, box.Handle(update{
		Value: 1,
		State: 3,
		Count: 1,
	}))
	require.Zero(t, state)
	require.Empty(t, updates)
	require.Empty(t, box.pending)

	require.Nil(t, box.Handle(update{
		Value: 1,
		State: 4,
		Count: 1,
	}))
	require.Equal(t, 4, state)
	require.Equal(t, []update{{1, 4, 1, nil}}, updates)
	require.Empty(t, box.pending)
	updates = nil

	require.Nil(t, box.Handle(update{
		Value: 1,
		State: 6,
		Count: 1,
	}))
	require.Equal(t, 4, state)
	require.Empty(t, updates)
	require.Equal(t, []update{{1, 6, 1, nil}}, box.pending)

	require.Nil(t, box.Handle(update{
		Value: 2,
		State: 5,
		Count: 1,
	}))
	require.Equal(t, 6, state)
	require.Equal(t, []update{{2, 5, 1, nil}, {1, 6, 1, nil}}, updates)
	updates = nil

	require.Nil(t, box.Handle(update{
		Value: 3,
		State: 8,
		Count: 1,
	}))
	require.Equal(t, 6, state)
	require.Empty(t, updates)
	require.Equal(t, []update{{3, 8, 1, nil}}, box.pending)
	<-box.gapTimeout.C

	require.Equal(t, []gap{{7, 7}}, box.gaps.gaps)
	box.EnableRecoverMode()
	require.True(t, box.recovering)
	require.False(t, box.gaps.Has())
}
