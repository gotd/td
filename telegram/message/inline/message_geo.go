package inline

import (
	"time"

	"github.com/nnqq/td/telegram/message/markup"
	"github.com/nnqq/td/tg"
)

// MessageMediaGeoBuilder is a builder of inline result geo message.
type MessageMediaGeoBuilder struct {
	message *tg.InputBotInlineMessageMediaGeo
}

// nolint:unparam
func (b *MessageMediaGeoBuilder) apply() (tg.InputBotInlineMessageClass, error) {
	r := *b.message
	return &r, nil
}

// MessageGeo creates new message geo option builder.
func MessageGeo(point tg.InputGeoPointClass) *MessageMediaGeoBuilder {
	return &MessageMediaGeoBuilder{
		message: &tg.InputBotInlineMessageMediaGeo{
			GeoPoint: point,
		},
	}
}

// Heading sets for live locations¹, a direction in which the location moves, in degrees; 1-360.
//
// Links:
//  1) https://core.telegram.org/api/live-location
func (b *MessageMediaGeoBuilder) Heading(heading int) *MessageMediaGeoBuilder {
	b.message.Heading = heading
	return b
}

// Period sets validity period.
func (b *MessageMediaGeoBuilder) Period(dur time.Duration) *MessageMediaGeoBuilder {
	return b.PeriodSeconds(int(dur.Seconds()))
}

// PeriodSeconds sets validity period in seconds.
func (b *MessageMediaGeoBuilder) PeriodSeconds(period int) *MessageMediaGeoBuilder {
	b.message.Period = period
	return b
}

// ProximityNotificationRadius sets for live locations¹, a maximum distance to another chat member for proximity
// alerts, in meters (0-100000)
//
// Links:
//  1) https://core.telegram.org/api/live-location
func (b *MessageMediaGeoBuilder) ProximityNotificationRadius(radius int) *MessageMediaGeoBuilder {
	b.message.ProximityNotificationRadius = radius
	return b
}

// Markup sets reply markup for sending bot buttons.
// NB: markup will not be used, if you send multiple media attachments.
func (b *MessageMediaGeoBuilder) Markup(m tg.ReplyMarkupClass) *MessageMediaGeoBuilder {
	b.message.ReplyMarkup = m
	return b
}

// Row sets single row keyboard markup  for sending bot buttons.
// NB: markup will not be used, if you send multiple media attachments.
func (b *MessageMediaGeoBuilder) Row(buttons ...tg.KeyboardButtonClass) *MessageMediaGeoBuilder {
	return b.Markup(markup.InlineRow(buttons...))
}
