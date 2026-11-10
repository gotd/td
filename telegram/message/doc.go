// Package message contains some useful utilities for creating Telegram messages.
//
// # Widgets
//
// Bots can send interactive HTML widgets, rendered by clients inside the chat
// bubble and sized to their content:
//
//	sender.To(peer).Media(ctx, message.Widget(html, styling.Plain("caption")).
//		PublicData(`{"count":0}`).
//		AllowQueries(true))
//
// A widget sent with AllowQueries can send queries to the bot; answer them with
// Sender.AnswerWidgetQuery. Data of a sent widget is edited with
// EditMessageBuilder.WidgetData (Sender.EditInlineWidgetData for inline
// messages). A widget sent with Shared(true) gets an identifier that can be
// sent again with SharedWidget(id).ChatData(data) and updated for all
// messages at once with Sender.EditSharedWidgetData. See package
// telegram/message/widget for limits and for decoding received widgets.
package message
