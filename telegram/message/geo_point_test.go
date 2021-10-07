package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func TestGeoPoint(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	point := tg.InputGeoPoint{
		Lat:            11,
		Long:           12,
		AccuracyRadius: 10,
	}

	expectSendMedia(t, &tg.InputMediaGeoPoint{
		GeoPoint: &point,
	}, mock)

	_, err := sender.Self().Media(ctx, GeoPoint(11, 12, 10))
	require.NoError(t, err)
}
