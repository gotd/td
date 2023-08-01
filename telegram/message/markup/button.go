package markup

import "github.com/gotd/td/tg"

// Row creates keyboard row.
func Row(buttons ...tg.KeyboardButtonClass) tg.KeyboardButtonRow {
	return tg.KeyboardButtonRow{
		Buttons: buttons,
	}
}

// Button creates new plain text button.
func Button(text string) *tg.KeyboardButton {
	return &tg.KeyboardButton{
		Text: text,
	}
}

// URL creates new URL button.
func URL(text, url string) *tg.KeyboardButtonURL {
	return &tg.KeyboardButtonURL{
		Text: text,
		URL:  url,
	}
}

// Callback creates new callback button.
func Callback(text string, data []byte) *tg.KeyboardButtonCallback {
	return &tg.KeyboardButtonCallback{
		Text: text,
		Data: data,
	}
}

// RequestPhone creates button to request a user's phone number.
func RequestPhone(text string) *tg.KeyboardButtonRequestPhone {
	return &tg.KeyboardButtonRequestPhone{
		Text: text,
	}
}

// RequestGeoLocation creates button to request a user's geo location.
func RequestGeoLocation(text string) *tg.KeyboardButtonRequestGeoLocation {
	return &tg.KeyboardButtonRequestGeoLocation{
		Text: text,
	}
}

// SwitchInline creates button to force a user to switch to inline mode.
// Pressing the button will prompt the user to select one of their chats, open that chat and insert the bot‘s username
// and the specified inline query in the input field.
//
// If samePeer set, pressing the button will insert the bot‘s
// username and the specified inline query in the current chat's input field.
func SwitchInline(text, query string, samePeer bool) *tg.KeyboardButtonSwitchInline {
	return &tg.KeyboardButtonSwitchInline{
		SamePeer: samePeer,
		Text:     text,
		Query:    query,
	}
}

// Game creates button to start a game.
func Game(text string) *tg.KeyboardButtonGame {
	return &tg.KeyboardButtonGame{
		Text: text,
	}
}

// Buy creates button to buy a product.
func Buy(text string) *tg.KeyboardButtonBuy {
	return &tg.KeyboardButtonBuy{
		Text: text,
	}
}

// InputURLAuth creates button to request a user to authorize via URL using Seamless Telegram Login.
// Can only be sent or received as part of an inline keyboard, use URLAuth for reply keyboards.
func InputURLAuth(requestWriteAccess bool, text, fwdText, url string, bot tg.InputUserClass) *tg.InputKeyboardButtonURLAuth {
	return &tg.InputKeyboardButtonURLAuth{
		RequestWriteAccess: requestWriteAccess,
		Text:               text,
		FwdText:            fwdText,
		URL:                url,
		Bot:                bot,
	}
}

// URLAuth creates button to request a user to authorize via URL using Seamless Telegram Login.
// Can only be sent or received as part of a reply keyboard, use InputURLAuth for inline keyboards.
func URLAuth(text, url string, buttonID int, fwdText string) *tg.KeyboardButtonURLAuth {
	return &tg.KeyboardButtonURLAuth{
		Text:     text,
		URL:      url,
		ButtonID: buttonID,
		FwdText:  fwdText,
	}
}

// RequestPoll creates button that allows the user to create and send a poll when pressed.
// Available only in private.
func RequestPoll(text string, quiz bool) *tg.KeyboardButtonRequestPoll {
	return &tg.KeyboardButtonRequestPoll{
		Text: text,
		Quiz: quiz,
	}
}

// InputUserProfile creates button that links directly to a user profile.
// Can only be sent or received as part of an inline keyboard, use UserProfile for reply keyboards.
func InputUserProfile(text string, user tg.InputUserClass) *tg.InputKeyboardButtonUserProfile {
	return &tg.InputKeyboardButtonUserProfile{
		Text:   text,
		UserID: user,
	}
}

// UserProfile creates button that links directly to a user profile.
// Can only be sent or received as part of a reply keyboard, use InputUserProfile for inline keyboards.
func UserProfile(text string, userID int64) *tg.KeyboardButtonUserProfile {
	return &tg.KeyboardButtonUserProfile{
		Text:   text,
		UserID: userID,
	}
}

// WebView creates button to open a bot web app using messages.requestWebView, sending over user information after
// user confirmation.
// Can only be sent or received as part of an inline keyboard, use SimpleWebView for reply keyboards.
func WebView(text, url string) *tg.KeyboardButtonWebView {
	return &tg.KeyboardButtonWebView{
		Text: text,
		URL:  url,
	}
}

// SimpleWebView creates button to open a bot web app using messages.requestSimpleWebView, without sending user
// information to the web app.
// Can only be sent or received as part of a reply keyboard, use WebView for inline keyboards.
func SimpleWebView(text, url string) *tg.KeyboardButtonSimpleWebView {
	return &tg.KeyboardButtonSimpleWebView{
		Text: text,
		URL:  url,
	}
}

// RequestPeer creates button that prompts the user to select and share a peer with the bot using
// messages.sendBotRequestedPeer.
func RequestPeer(text string, buttonID int, peerType tg.RequestPeerTypeClass) *tg.KeyboardButtonRequestPeer {
	return &tg.KeyboardButtonRequestPeer{
		Text:     text,
		ButtonID: buttonID,
		PeerType: peerType,
	}
}
