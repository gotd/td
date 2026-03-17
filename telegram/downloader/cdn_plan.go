package downloader

import "github.com/go-faster/errors"

type cdnRequestRange struct {
	offset int64
	limit  int
}

const (
	cdnMinChunk = 4 * 1024
	cdnMaxChunk = 1024 * 1024
)

func largestCDNValidLimit(max int) int {
	for size := max; size >= cdnMinChunk; size -= cdnMinChunk {
		if cdnMaxChunk%size == 0 {
			return size
		}
	}
	return 0
}

func buildCDNRequestPlan(offset int64, limit int) ([]cdnRequestRange, error) {
	if limit <= 0 {
		return nil, errors.Errorf("invalid CDN limit %d", limit)
	}
	if offset < 0 {
		return nil, errors.Errorf("invalid CDN offset %d", offset)
	}
	if offset%cdnMinChunk != 0 {
		return nil, errors.Errorf("CDN offset %d must be divisible by %d", offset, cdnMinChunk)
	}
	if limit%cdnMinChunk != 0 {
		return nil, errors.Errorf("CDN limit %d must be divisible by %d", limit, cdnMinChunk)
	}

	remaining := limit
	current := offset
	plan := make([]cdnRequestRange, 0, 1+limit/cdnMaxChunk)
	for remaining > 0 {
		mbUsed := int(current % cdnMaxChunk)
		mbLeft := cdnMaxChunk - mbUsed
		maxForStep := remaining
		if maxForStep > mbLeft {
			maxForStep = mbLeft
		}
		// Step size is chosen from values allowed by CDN docs:
		// - divisible by 4KB
		// - divisor of 1MB.
		step := largestCDNValidLimit(maxForStep)
		if step == 0 {
			return nil, errors.Errorf("unable to build CDN request plan for offset=%d limit=%d", offset, limit)
		}
		plan = append(plan, cdnRequestRange{
			offset: current,
			limit:  step,
		})
		current += int64(step)
		remaining -= step
	}

	return plan, nil
}
