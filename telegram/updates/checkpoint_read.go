package updates

import "context"

// checkpointReadStorage classifies channel checkpoint read errors, including
// reads when resuming a worker. Setters retain their apply-path error handling.
type checkpointReadStorage struct{ StateStorage }

func (s checkpointReadStorage) GetChannelPts(ctx context.Context, userID, channelID int64) (pts int, found bool, err error) {
	pts, found, err = s.StateStorage.GetChannelPts(ctx, userID, channelID)
	if err != nil {
		err = &applyError{err}
	}
	return pts, found, err
}
