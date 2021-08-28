package updates

import "testing"

func TestState(t *testing.T) {
	newState(stateConfig{
		State: State{
			Pts:  1,
			Qts:  1,
			Date: 1,
			Seq:  1,
		},
	})
}
