package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestVenue(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	point := tg.InputGeoPoint{
		Lat:            11,
		Long:           12,
		AccuracyRadius: 10,
	}

	expectSendMedia(t, &tg.InputMediaVenue{
		Title:    "Test Venue",
		Address:  "Test Address",
		GeoPoint: &point,
	}, mock)

	_, err := sender.Self().Media(ctx, Venue(11, 12, 10, "Test Venue", "Test Address"))
	require.NoError(t, err)
}
