package markup

import "github.com/gotd/td/tg"

// StyleOption is a functional parameter that configures the tg.KeyboardButtonStyle
// of a button, allowing a custom background color and a custom emoji label.
//
// See https://core.telegram.org/api/bots/buttons#button-styles.
type StyleOption func(s *tg.KeyboardButtonStyle)

// applyStyle builds a tg.KeyboardButtonStyle from the given options.
func applyStyle(options []StyleOption) tg.KeyboardButtonStyle {
	var style tg.KeyboardButtonStyle
	for _, opt := range options {
		opt(&style)
	}
	return style
}

// StyleBgPrimary sets a dark blue background color, recommended for main actions.
func StyleBgPrimary() StyleOption {
	return func(s *tg.KeyboardButtonStyle) {
		s.BgPrimary = true
	}
}

// StyleBgDanger sets a red background color, recommended for destructive actions.
func StyleBgDanger() StyleOption {
	return func(s *tg.KeyboardButtonStyle) {
		s.BgDanger = true
	}
}

// StyleBgSuccess sets a green background color, recommended for positive actions.
func StyleBgSuccess() StyleOption {
	return func(s *tg.KeyboardButtonStyle) {
		s.BgSuccess = true
	}
}

// StyleIcon sets the ID of a custom emoji to be displayed before the button's label.
//
// See https://core.telegram.org/api/custom-emoji.
func StyleIcon(icon int64) StyleOption {
	return func(s *tg.KeyboardButtonStyle) {
		s.SetIcon(icon)
	}
}

// reply creates a reply keyboard button of given type.
func reply(text string, typ tg.ButtonTypeClass, style []StyleOption) tg.KeyboardButton {
	return tg.KeyboardButton{
		Text:  text,
		Type:  typ,
		Style: applyStyle(style),
	}
}

// inline creates an inline keyboard button of given type.
func inline(text string, typ tg.InlineButtonTypeClass, style []StyleOption) tg.KeyboardInlineButton {
	return tg.KeyboardInlineButton{
		Text:  text,
		Type:  typ,
		Style: applyStyle(style),
	}
}

// Row creates reply keyboard row.
func Row(buttons ...tg.KeyboardButton) tg.KeyboardButtonRow {
	return tg.KeyboardButtonRow{
		Buttons: buttons,
	}
}

// InlineButtonRow creates inline keyboard row.
func InlineButtonRow(buttons ...tg.KeyboardInlineButton) tg.KeyboardInlineButtonRow {
	return tg.KeyboardInlineButtonRow{
		Buttons: buttons,
	}
}

// Button creates new plain text button.
func Button(text string, style ...StyleOption) tg.KeyboardButton {
	return reply(text, &tg.ButtonTypeDefault{}, style)
}

// RequestPhone creates button to request a user's phone number.
func RequestPhone(text string, style ...StyleOption) tg.KeyboardButton {
	return reply(text, &tg.ButtonTypeRequestPhone{}, style)
}

// RequestGeoLocation creates button to request a user's geo location.
func RequestGeoLocation(text string, style ...StyleOption) tg.KeyboardButton {
	return reply(text, &tg.ButtonTypeRequestGeoLocation{}, style)
}

// RequestPoll creates button that allows the user to create and send a poll when pressed.
// Available only in private.
func RequestPoll(text string, quiz bool, style ...StyleOption) tg.KeyboardButton {
	return reply(text, &tg.ButtonTypeRequestPoll{Quiz: quiz}, style)
}

// RequestPeer creates button that prompts the user to select and share a peer with the bot using
// messages.sendBotRequestedPeer.
func RequestPeer(text string, buttonID int, peerType tg.RequestPeerTypeClass, style ...StyleOption) tg.KeyboardButton {
	return reply(text, &tg.ButtonTypeRequestPeer{
		ButtonID: buttonID,
		PeerType: peerType,
	}, style)
}

// SimpleWebView creates button to open a bot web app using messages.requestSimpleWebView, without sending user
// information to the web app.
// Can only be sent or received as part of a reply keyboard, use WebView for inline keyboards.
func SimpleWebView(text, url string, style ...StyleOption) tg.KeyboardButton {
	return reply(text, &tg.ButtonTypeSimpleWebView{URL: url}, style)
}

// URL creates new URL button.
func URL(text, url string, style ...StyleOption) tg.KeyboardInlineButton {
	return inline(text, &tg.InlineButtonTypeURL{URL: url}, style)
}

// Callback creates new callback button.
func Callback(text string, data []byte, style ...StyleOption) tg.KeyboardInlineButton {
	return inline(text, &tg.InlineButtonTypeCallback{Data: data}, style)
}

// SwitchInline creates button to force a user to switch to inline mode.
// Pressing the button will prompt the user to select one of their chats, open that chat and insert the bot‘s username
// and the specified inline query in the input field.
//
// If samePeer set, pressing the button will insert the bot‘s
// username and the specified inline query in the current chat's input field.
func SwitchInline(text, query string, samePeer bool, style ...StyleOption) tg.KeyboardInlineButton {
	return inline(text, &tg.InlineButtonTypeSwitchInline{
		SamePeer: samePeer,
		Query:    query,
	}, style)
}

// Game creates button to start a game.
func Game(text string, style ...StyleOption) tg.KeyboardInlineButton {
	return inline(text, &tg.InlineButtonTypeGame{}, style)
}

// Buy creates button to buy a product.
func Buy(text string, style ...StyleOption) tg.KeyboardInlineButton {
	return inline(text, &tg.InlineButtonTypeBuy{}, style)
}

// InputURLAuth creates button to request a user to authorize via URL using Seamless Telegram Login.
// Use it to send a button, use URLAuth to handle a received one.
func InputURLAuth(
	requestWriteAccess bool, text, fwdText, url string, bot tg.InputUserClass, style ...StyleOption,
) tg.KeyboardInlineButton {
	return inline(text, &tg.InputInlineButtonTypeURLAuth{
		RequestWriteAccess: requestWriteAccess,
		FwdText:            fwdText,
		URL:                url,
		Bot:                bot,
	}, style)
}

// URLAuth creates button to request a user to authorize via URL using Seamless Telegram Login.
// Use InputURLAuth to send such a button.
func URLAuth(text, url string, buttonID int, fwdText string, style ...StyleOption) tg.KeyboardInlineButton {
	return inline(text, &tg.InlineButtonTypeURLAuth{
		FwdText:  fwdText,
		URL:      url,
		ButtonID: buttonID,
	}, style)
}

// InputUserProfile creates button that links directly to a user profile.
// Use it to send a button, use UserProfile to handle a received one.
func InputUserProfile(text string, user tg.InputUserClass, style ...StyleOption) tg.KeyboardInlineButton {
	return inline(text, &tg.InputInlineButtonTypeUserProfile{UserID: user}, style)
}

// UserProfile creates button that links directly to a user profile.
// Use InputUserProfile to send such a button.
func UserProfile(text string, userID int64, style ...StyleOption) tg.KeyboardInlineButton {
	return inline(text, &tg.InlineButtonTypeUserProfile{UserID: userID}, style)
}

// WebView creates button to open a bot web app using messages.requestWebView, sending over user information after
// user confirmation.
// Can only be sent or received as part of an inline keyboard, use SimpleWebView for reply keyboards.
func WebView(text, url string, style ...StyleOption) tg.KeyboardInlineButton {
	return inline(text, &tg.InlineButtonTypeWebView{URL: url}, style)
}
