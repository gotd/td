package message

import (
	"time"

	"github.com/nnqq/td/telegram/message/markup"
	"github.com/nnqq/td/telegram/message/peer"
	"github.com/nnqq/td/tg"
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

	// The message ID to which this message will reply to.
	replyToMsgID int
	// Reply markup for sending bot buttons.
	replyMarkup tg.ReplyMarkupClass
	// Scheduled message date for scheduled messages.
	scheduleDate int
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
	r := b.copy()
	r.silent = true
	return r
}

// Background sets flag to send this message as background message.
func (b *Builder) Background() *Builder {
	r := b.copy()
	r.background = true
	return r
}

// Clear sets flag to clear the draft field.
func (b *Builder) Clear() *Builder {
	r := b.copy()
	r.clearDraft = true
	return r
}

// Reply sets message ID to reply.
func (b *Builder) Reply(id int) *Builder {
	r := b.copy()
	r.replyToMsgID = id
	return r
}

// ReplyMsg sets message to reply.
func (b *Builder) ReplyMsg(msg tg.MessageClass) *Builder {
	return b.Reply(msg.GetID())
}

// ScheduleTS sets scheduled message timestamp for scheduled messages.
func (b *Builder) ScheduleTS(date int) *Builder {
	r := b.copy()
	r.scheduleDate = date
	return r
}

// Schedule sets scheduled message date for scheduled messages.
func (b *Builder) Schedule(date time.Time) *Builder {
	return b.ScheduleTS(int(date.Unix()))
}

// NoWebpage sets flag to disable generation of the webpage preview.
func (b *Builder) NoWebpage() *Builder {
	r := b.copy()
	r.noWebpage = true
	return r
}

// Markup sets reply markup for sending bot buttons.
// NB: markup will not be used, if you send multiple media attachments.
func (b *Builder) Markup(m tg.ReplyMarkupClass) *Builder {
	r := b.copy()
	r.replyMarkup = m
	return r
}

// Row sets single row keyboard markup  for sending bot buttons.
// NB: markup will not be used, if you send multiple media attachments.
func (b *Builder) Row(buttons ...tg.KeyboardButtonClass) *Builder {
	return b.Markup(markup.InlineRow(buttons...))
}
