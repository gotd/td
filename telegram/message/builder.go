package message

import (
	"time"

	"github.com/gotd/td/telegram/message/markup"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

type peerPromise = peer.Promise

// CloneBuilder returns copy of message Builder inside RequestBuilder.
func (b *RequestBuilder) CloneBuilder() *Builder {
	return b.Builder.copy()
}

// Builder is a message builder.
type Builder struct {
	// Sender to use.
	sender *Sender
	// The destination where the message will be sent.
	peer peerPromise

	// Set this flag to disable generation of the webpage preview.
	noWebpage bool
	// Send this message silently (no notifications for the receivers).
	silent bool
	// Send this message as background message.
	background bool
	// Clear the draft field.
	clearDraft bool
	// noForwards whether that sent message cannot be forwarded.
	noForwards bool

	// The message ID to which this message will reply to.
	replyToMsgID int
	// Reply markup for sending bot buttons.
	replyMarkup tg.ReplyMarkupClass
	// Scheduled message date for scheduled messages.
	scheduleDate int

	// sendAs sets peer to send message as it.
	sendAs tg.InputPeerClass
}

func (b *Builder) copy() *Builder {
	if b == nil {
		return nil
	}

	r := *b
	return &r
}

// Silent sets flag to send this message silently (no notifications for the receivers).
func (b *Builder) Silent() *Builder {
	b.silent = true
	return b
}

// Background sets flag to send this message as background message.
func (b *Builder) Background() *Builder {
	b.background = true
	return b
}

// Clear sets flag to clear the draft field.
func (b *Builder) Clear() *Builder {
	b.clearDraft = true
	return b
}

// Reply sets message ID to reply.
func (b *Builder) Reply(id int) *Builder {
	b.replyToMsgID = id
	return b
}

// ReplyMsg sets message to reply.
func (b *Builder) ReplyMsg(msg tg.MessageClass) *Builder {
	return b.Reply(msg.GetID())
}

// ScheduleTS sets scheduled message timestamp for scheduled messages.
func (b *Builder) ScheduleTS(date int) *Builder {
	b.scheduleDate = date
	return b
}

// Schedule sets scheduled message date for scheduled messages.
func (b *Builder) Schedule(date time.Time) *Builder {
	return b.ScheduleTS(int(date.Unix()))
}

// NoWebpage sets flag to disable generation of the webpage preview.
func (b *Builder) NoWebpage() *Builder {
	b.noWebpage = true
	return b
}

// NoForwards whether that sent message cannot be forwarded.
//
// See https://telegram.org/blog/protected-content-delete-by-date-and-more#protected-content-in-groups-and-channels.
func (b *Builder) NoForwards() *Builder {
	b.noForwards = true
	return b
}

// Markup sets reply markup for sending bot buttons.
//
// NB: markup will not be used, if you send multiple media attachments.
func (b *Builder) Markup(m tg.ReplyMarkupClass) *Builder {
	b.replyMarkup = m
	return b
}

// Row sets single row keyboard markup  for sending bot buttons.
//
// NB: markup will not be used, if you send multiple media attachments.
func (b *Builder) Row(buttons ...tg.KeyboardButtonClass) *Builder {
	return b.Markup(markup.InlineRow(buttons...))
}

// SendAs sets peer to send as.
//
// See https://telegram.org/blog/protected-content-delete-by-date-and-more#anonymous-posting-in-public-groups.
func (b *Builder) SendAs(p tg.InputPeerClass) *Builder {
	b.sendAs = p
	return b
}
