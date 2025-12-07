package message

import "github.com/gotd/td/tg"

// Venue adds venue attachment.
// NB: parameter accuracy may be zero and will not be used.
func Venue(lat, long float64, accuracy int, title, address string, caption ...StyledTextOption) MediaOption {
	return Media(&tg.InputMediaVenue{
		Title:   title,
		Address: address,
		GeoPoint: &tg.InputGeoPoint{
			Lat:            lat,
			Long:           long,
			AccuracyRadius: accuracy,
		},
	}, caption...)
}
