package hook

import (
	"context"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/assert"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type recordedAffected struct {
	called    bool
	channelID int64
	pts       int
	ptsCount  int
}

func (r *recordedAffected) HandleAffected(_ context.Context, channelID int64, pts, ptsCount int) error {
	r.called = true
	r.channelID = channelID
	r.pts = pts
	r.ptsCount = ptsCount
	return nil
}

func invoke(t *testing.T, rec *recordedAffected, input bin.Encoder, output bin.Decoder) error {
	t.Helper()
	return AffectedHook(rec).Handle(telegram.InvokeFunc(
		func(context.Context, bin.Encoder, bin.Decoder) error { return nil },
	)).Invoke(context.Background(), input, output)
}

func TestAffectedHook(t *testing.T) {
	t.Run("CommonByPeer", func(t *testing.T) {
		rec := &recordedAffected{}
		input := &tg.MessagesReadHistoryRequest{Peer: &tg.InputPeerUser{UserID: 1}}
		assert.NoError(t, invoke(t, rec, input, &tg.MessagesAffectedMessages{Pts: 5, PtsCount: 1}))
		assert.True(t, rec.called)
		assert.Equal(t, int64(0), rec.channelID, "user peer routes to common")
		assert.Equal(t, 5, rec.pts)
		assert.Equal(t, 1, rec.ptsCount)
	})

	t.Run("CommonNoPeer", func(t *testing.T) {
		rec := &recordedAffected{}
		input := &tg.MessagesDeleteMessagesRequest{ID: []int{1, 2}}
		assert.NoError(t, invoke(t, rec, input, &tg.MessagesAffectedMessages{Pts: 9, PtsCount: 2}))
		assert.True(t, rec.called)
		assert.Equal(t, int64(0), rec.channelID)
		assert.Equal(t, 9, rec.pts)
	})

	t.Run("ChannelByInputChannel", func(t *testing.T) {
		rec := &recordedAffected{}
		input := &tg.ChannelsDeleteMessagesRequest{
			Channel: &tg.InputChannel{ChannelID: 77, AccessHash: 1},
			ID:      []int{3},
		}
		assert.NoError(t, invoke(t, rec, input, &tg.MessagesAffectedMessages{Pts: 4, PtsCount: 1}))
		assert.True(t, rec.called)
		assert.Equal(t, int64(77), rec.channelID, "InputChannel routes to that channel")
	})

	t.Run("ChannelByPeer", func(t *testing.T) {
		rec := &recordedAffected{}
		input := &tg.MessagesReadMentionsRequest{Peer: &tg.InputPeerChannel{ChannelID: 88, AccessHash: 1}}
		assert.NoError(t, invoke(t, rec, input, &tg.MessagesAffectedHistory{Pts: 7, PtsCount: 1}))
		assert.True(t, rec.called)
		assert.Equal(t, int64(88), rec.channelID, "channel peer routes to that channel")
		assert.Equal(t, 7, rec.pts)
	})

	t.Run("NotAffectedResult", func(t *testing.T) {
		rec := &recordedAffected{}
		assert.NoError(t, invoke(t, rec, &tg.MessagesReadHistoryRequest{}, &tg.UpdatesBox{
			Updates: &tg.UpdateShortMessage{ID: 1},
		}))
		assert.False(t, rec.called, "non-affected results must be ignored")
	})

	t.Run("InvokeErrorSkipsHook", func(t *testing.T) {
		rec := &recordedAffected{}
		boom := errors.New("boom")
		err := AffectedHook(rec).Handle(telegram.InvokeFunc(
			func(context.Context, bin.Encoder, bin.Decoder) error { return boom },
		)).Invoke(context.Background(), &tg.MessagesReadHistoryRequest{}, &tg.MessagesAffectedMessages{Pts: 5, PtsCount: 1})
		assert.ErrorIs(t, err, boom)
		assert.False(t, rec.called, "hook must not run when the RPC failed")
	})
}
