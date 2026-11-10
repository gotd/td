package message

import (
	"context"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/message/widget"
	"github.com/gotd/td/tg"
)

// WidgetBuilder is an HTML widget media option.
//
// Widgets can be sent only by bots. See package telegram/message/widget for
// details and limits.
type WidgetBuilder struct {
	input   tg.InputMediaWidget
	caption []StyledTextOption
}

// Widget creates a new HTML widget media option.
//
// The html must be 1-32768 bytes of UTF-8, which is checked before sending
// (see widget.ValidateHTML). Telegram also requires it to be at most 4096
// bytes when compressed by zlib with the maximum compression level; that is
// enforced by the server only (use widget.ValidateCompressed for an
// approximate early check). Use Builder.InvertMedia to show the caption above
// the widget.
func Widget(html string, caption ...StyledTextOption) *WidgetBuilder {
	return &WidgetBuilder{
		input:   tg.InputMediaWidget{HTML: html},
		caption: caption,
	}
}

// Size sets widget width and height. Zero width means that the widget
// occupies the whole width of the message bubble, zero height means that the
// height is not fixed (the widget is sized to its content). Defaults to 0x0.
func (w *WidgetBuilder) Size(width, height int) *WidgetBuilder {
	w.input.Width = width
	w.input.Height = height
	return w
}

// PublicData sets public data of the widget, which is also available in copies
// and forwards of the message.
func (w *WidgetBuilder) PublicData(data string) *WidgetBuilder {
	w.input.PublicData = data
	return w
}

// ChatData sets data of the widget, which is specific to the chat.
func (w *WidgetBuilder) ChatData(data string) *WidgetBuilder {
	w.input.ChatData = data
	return w
}

// AllowQueries allows the widget to send queries to the bot.
//
// Queries arrive as updateBotWidgetQuery (updateInlineBotWidgetQuery for inline
// messages, updateBotWidgetCopyQuery for copies of shared widgets) and are
// answered with Sender.AnswerWidgetQuery.
func (w *WidgetBuilder) AllowQueries(allow bool) *WidgetBuilder {
	w.input.AllowQueries = allow
	return w
}

// Shared makes the widget shared: it can be sent again by its identifier
// (SharedWidget) and its public data can be edited at once for all messages
// with Sender.EditSharedWidgetData. The identifier is returned in
// messageMediaWidget.widget_id of the sent message.
func (w *WidgetBuilder) Shared(shared bool) *WidgetBuilder {
	w.input.SharedWidget = shared
	return w
}

func (w *WidgetBuilder) apply(ctx context.Context, b *multiMediaBuilder) error {
	if err := widget.ValidateSize(w.input.Width, w.input.Height); err != nil {
		return err
	}
	if err := widget.ValidateHTML(w.input.HTML); err != nil {
		return err
	}
	input := w.input
	return Media(&input, w.caption...).apply(ctx, b)
}

// SharedWidgetBuilder is a media option that sends an existing shared
// widget by its identifier.
type SharedWidgetBuilder struct {
	input   tg.InputMediaWidgetID
	caption []StyledTextOption
}

// SharedWidget creates media option that sends an existing shared widget by
// its identifier, with optional caption.
func SharedWidget(id int64, caption ...StyledTextOption) *SharedWidgetBuilder {
	return &SharedWidgetBuilder{
		input:   tg.InputMediaWidgetID{WidgetID: id},
		caption: caption,
	}
}

// ChatData sets data of the widget, which is specific to the chat.
func (w *SharedWidgetBuilder) ChatData(data string) *SharedWidgetBuilder {
	w.input.ChatData = data
	return w
}

func (w *SharedWidgetBuilder) apply(ctx context.Context, b *multiMediaBuilder) error {
	if w.input.WidgetID == 0 {
		return widget.ErrInvalidWidgetID
	}
	input := w.input
	return Media(&input, w.caption...).apply(ctx, b)
}

// Widget sends an HTML widget. Use Media with the Widget option to set the
// widget size, data and flags.
func (b *Builder) Widget(ctx context.Context, html string, caption ...StyledTextOption) (tg.UpdatesClass, error) {
	return b.Media(ctx, Widget(html, caption...))
}

// SharedWidget sends an existing shared widget by its identifier. Use Media
// with the SharedWidget option to set chat data.
func (b *Builder) SharedWidget(ctx context.Context, id int64, caption ...StyledTextOption) (tg.UpdatesClass, error) {
	return b.Media(ctx, SharedWidget(id, caption...))
}

// widgetData is a set of widget data fields to edit.
type widgetData struct {
	publicData    string
	setPublicData bool
	chatData      string
	setChatData   bool
}

// WidgetDataOption is an option of widget data editing.
type WidgetDataOption func(d *widgetData)

// WidgetPublicData sets new public data of the widget. Empty string clears it.
func WidgetPublicData(data string) WidgetDataOption {
	return func(d *widgetData) {
		d.publicData = data
		d.setPublicData = true
	}
}

// WidgetChatData sets new data of the widget, which is specific to the chat.
// Empty string clears it.
func WidgetChatData(data string) WidgetDataOption {
	return func(d *widgetData) {
		d.chatData = data
		d.setChatData = true
	}
}

// ErrNoWidgetData is returned when widget data editing is requested without
// WidgetPublicData or WidgetChatData options.
//
// Unlike tdlib, which returns Ok without sending a request in this case, the
// helpers report it as an error, as it is most likely a programming mistake.
var ErrNoWidgetData = errors.New("no widget data to edit")

func buildWidgetData(opts []WidgetDataOption) (widgetData, error) {
	var d widgetData
	for _, opt := range opts {
		opt(&d)
	}
	if !d.setPublicData && !d.setChatData {
		return d, ErrNoWidgetData
	}
	return d, nil
}

// WidgetData edits data of the widget in the message sent by the bot
// (messages.editMessageWidgetData). Only fields set by options are changed;
// without options ErrNoWidgetData is returned and nothing is sent.
//
//	sender.To(peer).Edit(msgID).WidgetData(ctx, message.WidgetPublicData(`{"n":1}`))
func (b *EditMessageBuilder) WidgetData(ctx context.Context, opts ...WidgetDataOption) (tg.UpdatesClass, error) {
	d, err := buildWidgetData(opts)
	if err != nil {
		return nil, err
	}
	p, err := b.builder.peer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "peer")
	}

	req := &tg.MessagesEditMessageWidgetDataRequest{
		Peer: p,
		ID:   b.id,
	}
	if d.setPublicData {
		req.SetPublicData(d.publicData)
	}
	if d.setChatData {
		req.SetChatData(d.chatData)
	}

	upd, err := b.builder.sender.raw.MessagesEditMessageWidgetData(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "edit message widget data")
	}

	return upd, nil
}

// EditInlineWidgetData edits data of the widget in an inline message sent via
// the bot (messages.editInlineBotMessageWidgetData). Only fields set by options
// are changed; without options ErrNoWidgetData is returned and nothing is
// sent.
//
// As with other inline message edits, Telegram may require the request to be
// sent to the DC of the inline message (the DCID field of the identifier); use
// a Sender created over telegram.Client.DC for that.
func (s *Sender) EditInlineWidgetData(
	ctx context.Context, id tg.InputBotInlineMessageIDClass, opts ...WidgetDataOption,
) error {
	d, err := buildWidgetData(opts)
	if err != nil {
		return err
	}

	req := &tg.MessagesEditInlineBotMessageWidgetDataRequest{ID: id}
	if d.setPublicData {
		req.SetPublicData(d.publicData)
	}
	if d.setChatData {
		req.SetChatData(d.chatData)
	}

	if _, err := s.raw.MessagesEditInlineBotMessageWidgetData(ctx, req); err != nil {
		return errors.Wrap(err, "edit inline message widget data")
	}
	return nil
}

const (
	minWidgetCacheTime = 5 * time.Second
	maxWidgetCacheTime = time.Hour
)

// ErrInvalidWidgetCacheTime is returned when shared widget cache time is not
// zero and is out of the 5s-1h range.
var ErrInvalidWidgetCacheTime = errors.New("invalid widget cache time")

// EditSharedWidgetData edits public data of the shared widget created by the
// bot, for every message showing it (messages.editWidgetData).
//
// The cacheTime is the time during which clients may cache the public data,
// 5s-1h, in whole seconds; zero keeps the current value (30s by default).
func (s *Sender) EditSharedWidgetData(
	ctx context.Context, widgetID int64, publicData string, cacheTime time.Duration,
) error {
	if widgetID == 0 {
		return widget.ErrInvalidWidgetID
	}
	if cacheTime != 0 && (cacheTime < minWidgetCacheTime || cacheTime > maxWidgetCacheTime) {
		return errors.Wrapf(ErrInvalidWidgetCacheTime, "%s", cacheTime)
	}

	req := &tg.MessagesEditWidgetDataRequest{
		WidgetID:   widgetID,
		PublicData: publicData,
	}
	if cacheTime != 0 {
		req.SetCacheTime(int(cacheTime / time.Second))
	}

	if _, err := s.raw.MessagesEditWidgetData(ctx, req); err != nil {
		return errors.Wrap(err, "edit shared widget data")
	}
	return nil
}

// AnswerWidgetQuery sets the result of a widget query
// (messages.setBotWidgetQueryAnswer). The queryID is QueryID of
// updateBotWidgetQuery, updateInlineBotWidgetQuery or updateBotWidgetCopyQuery;
// result is the data returned to the widget, empty for no data.
func (s *Sender) AnswerWidgetQuery(ctx context.Context, queryID int64, result string) error {
	if _, err := s.raw.MessagesSetBotWidgetQueryAnswer(ctx, &tg.MessagesSetBotWidgetQueryAnswerRequest{
		QueryID: queryID,
		Result:  result,
	}); err != nil {
		return errors.Wrap(err, "set widget query answer")
	}
	return nil
}
