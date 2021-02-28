package message

import (
	"context"
	"time"

	"github.com/gotd/td/tg"
)

type peerPromise func(ctx context.Context) (tg.InputPeerClass, error)

// Builder is a message builder.
type Builder struct {
	// Sender to use.
	sender Sender
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

	// Attached media.
	media []tg.InputSingleMedia
	// The message ID to which this message will reply to.
	replyToMsgID int
	// Reply markup for sending bot buttons.
	replyMarkup tg.ReplyMarkupClass
	// Scheduled message date for scheduled messages.
	scheduleDate int
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

// ClearDraft sets flag to clear the draft field.
func (b *Builder) ClearDraft() *Builder {
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

// Schedule sets scheduled message date for scheduled messages.
func (b *Builder) Schedule(date time.Time) *Builder {
	b.scheduleDate = int(date.Unix())
	return b
}

// Markup sets reply markup for sending bot buttons.
// NB: markup will not be used, if you send multiple media attachments.
func (b *Builder) Markup(markup tg.ReplyMarkupClass) *Builder {
	b.replyMarkup = markup
	return b
}
