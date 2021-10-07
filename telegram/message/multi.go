package message

import (
	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
)

func convertMessageMediaToInput(m tg.MessageMediaClass) (tg.InputMediaClass, error) {
	switch v := m.(type) {
	case *tg.MessageMediaPhoto: // messageMediaPhoto#695150d7
		photo, ok := v.Photo.AsNotEmpty()
		if !ok {
			return nil, xerrors.Errorf("unexpected type %T", v.Photo)
		}

		return &tg.InputMediaPhoto{
			ID:         photo.AsInput(),
			TTLSeconds: v.TTLSeconds,
		}, nil
	case *tg.MessageMediaGeo: // messageMediaGeo#56e0d474
		geo, ok := v.Geo.AsNotEmpty()
		if !ok {
			return nil, xerrors.Errorf("unexpected type %T", v.Geo)
		}

		r := new(tg.InputGeoPoint)
		r.FillFrom(geo)
		return &tg.InputMediaGeoPoint{
			GeoPoint: r,
		}, nil
	case *tg.MessageMediaContact: // messageMediaContact#cbf24940
		r := new(tg.InputMediaContact)
		r.FillFrom(v)
		return r, nil
	case *tg.MessageMediaDocument: // messageMediaDocument#9cb070d7
		document, ok := v.Document.AsNotEmpty()
		if !ok {
			return nil, xerrors.Errorf("unexpected type %T", v.Document)
		}

		return &tg.InputMediaDocument{
			ID:         document.AsInput(),
			TTLSeconds: v.TTLSeconds,
		}, nil
	case *tg.MessageMediaVenue: // messageMediaVenue#2ec0533f
		geo, ok := v.Geo.AsNotEmpty()
		if !ok {
			return nil, xerrors.Errorf("unexpected type %T", v.Geo)
		}

		r := new(tg.InputGeoPoint)
		r.FillFrom(geo)
		return &tg.InputMediaVenue{
			GeoPoint:  r,
			Title:     v.Title,
			Address:   v.Address,
			Provider:  v.Provider,
			VenueID:   v.VenueID,
			VenueType: v.VenueType,
		}, nil
	case *tg.MessageMediaGame: // messageMediaGame#fdb19008
		r := new(tg.InputGameID)
		r.FillFrom(&v.Game)

		return &tg.InputMediaGame{
			ID: r,
		}, nil
	case *tg.MessageMediaGeoLive: // messageMediaGeoLive#b940c666
		geo, ok := v.Geo.AsNotEmpty()
		if !ok {
			return nil, xerrors.Errorf("unexpected type %T", v.Geo)
		}

		r := new(tg.InputGeoPoint)
		r.FillFrom(geo)
		return &tg.InputMediaGeoLive{
			GeoPoint:                    r,
			Heading:                     v.Heading,
			Period:                      v.Period,
			ProximityNotificationRadius: v.ProximityNotificationRadius,
		}, nil
	case *tg.MessageMediaDice: // messageMediaDice#3f7ee58b
		r := new(tg.InputMediaDice)
		r.FillFrom(v)
		return r, nil
	default:
		// messageMediaPoll#4bd6e798
		// messageMediaWebPage#a32dd600
		// messageMediaEmpty#3ded6320
		// messageMediaUnsupported#9f84f49e
		// messageMediaInvoice#84551347
		return nil, xerrors.Errorf("unexpected type %T", v)
	}
}
