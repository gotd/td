package updates

type gapCheckResult byte

const (
	_ gapCheckResult = iota
	gapApply
	gapIgnore
	gapRefetch
)

func checkGap(localState, remoteState, count int) gapCheckResult {
	if localState+count == remoteState {
		return gapApply
	}

	if localState+count > remoteState {
		return gapIgnore
	}

	return gapRefetch
}
