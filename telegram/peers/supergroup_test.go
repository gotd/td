package peers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func getTestSuperGroup() *tg.Channel {
	testChannel := getTestChannel()
	testChannel.Broadcast = false
	testChannel.Megagroup = true
	return testChannel
}

func TestSupergroup_ToggleSlowMode(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	ch := m.Channel(getTestSuperGroup())

	s, ok := ch.ToSupergroup()
	a.True(ok)

	mock.ExpectCall(&tg.ChannelsToggleSlowModeRequest{
		Channel: s.InputChannel(),
		Seconds: 1,
	}).ThenRPCErr(getTestError())
	a.Error(s.ToggleSlowMode(ctx, 1))

	mock.ExpectCall(&tg.ChannelsToggleSlowModeRequest{
		Channel: s.InputChannel(),
		Seconds: 1,
	}).ThenResult(&tg.Updates{})
	a.NoError(s.ToggleSlowMode(ctx, 1))

	mock.ExpectCall(&tg.ChannelsToggleSlowModeRequest{
		Channel: s.InputChannel(),
		Seconds: 0,
	}).ThenResult(&tg.Updates{})
	a.NoError(s.DisableSlowMode(ctx))
}

func TestSupergroup_SetStickerSet(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	mock, m := testManager(t)

	set := &tg.InputStickerSetShortName{ShortName: "gotd_stickers"}
	ch := m.Channel(getTestSuperGroup())

	s, ok := ch.ToSupergroup()
	a.True(ok)

	mock.ExpectCall(&tg.ChannelsSetStickersRequest{
		Channel:    s.InputChannel(),
		Stickerset: set,
	}).ThenRPCErr(getTestError())
	a.Error(s.SetStickerSet(ctx, set))

	mock.ExpectCall(&tg.ChannelsSetStickersRequest{
		Channel:    s.InputChannel(),
		Stickerset: set,
	}).ThenTrue()
	a.NoError(s.SetStickerSet(ctx, set))

	mock.ExpectCall(&tg.ChannelsSetStickersRequest{
		Channel:    s.InputChannel(),
		Stickerset: &tg.InputStickerSetEmpty{},
	}).ThenTrue()
	a.NoError(s.ResetStickerSet(ctx))
}
