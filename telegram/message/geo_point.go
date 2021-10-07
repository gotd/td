package message

import "github.com/nnqq/td/tg"

// GeoPoint adds geo point attachment.
// NB: parameter accuracy may be zero and will not be used.
func GeoPoint(lat, long float64, accuracy int, caption ...StyledTextOption) MediaOption {
	return Media(&tg.InputMediaGeoPoint{
		GeoPoint: &tg.InputGeoPoint{
			Lat:            lat,
			Long:           long,
			AccuracyRadius: accuracy,
		},
	}, caption...)
}
