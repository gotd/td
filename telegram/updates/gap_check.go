package updates

type gapCheckResult byte

const (
	_ gapCheckResult = iota
	gapApply
	gapIgnore
	gapRefetch
)

func checkGap(localState, remoteState, count int) gapCheckResult {
	// Temporary fix for handling qts updates gaps.
	if remoteState == 0 {
		return gapApply
	}

	if localState+count == remoteState {
		return gapApply
	}

	if localState+count > remoteState {
		return gapIgnore
	}

	return gapRefetch
}
