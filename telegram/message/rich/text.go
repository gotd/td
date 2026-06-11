package rich

import "github.com/gotd/td/tg"

// Empty returns an empty rich text node (textEmpty).
func Empty() *tg.TextEmpty {
	return &tg.TextEmpty{}
}

// Plain returns a plain (unstyled) rich text node (textPlain).
func Plain(s string) *tg.TextPlain {
	return &tg.TextPlain{Text: s}
}

// Concat concatenates the given rich text nodes (textConcat).
func Concat(texts ...tg.RichTextClass) *tg.TextConcat {
	return &tg.TextConcat{Texts: texts}
}

// Bold formats the given text as bold (textBold).
func Bold(texts ...tg.RichTextClass) *tg.TextBold {
	return &tg.TextBold{Text: join(texts)}
}

// Italic formats the given text as italic (textItalic).
func Italic(texts ...tg.RichTextClass) *tg.TextItalic {
	return &tg.TextItalic{Text: join(texts)}
}

// Underline formats the given text as underlined (textUnderline).
func Underline(texts ...tg.RichTextClass) *tg.TextUnderline {
	return &tg.TextUnderline{Text: join(texts)}
}

// Strike formats the given text as strikethrough (textStrike).
func Strike(texts ...tg.RichTextClass) *tg.TextStrike {
	return &tg.TextStrike{Text: join(texts)}
}

// Fixed formats the given text as fixed-width / monospace (textFixed).
func Fixed(texts ...tg.RichTextClass) *tg.TextFixed {
	return &tg.TextFixed{Text: join(texts)}
}

// Subscript formats the given text as subscript (textSubscript).
func Subscript(texts ...tg.RichTextClass) *tg.TextSubscript {
	return &tg.TextSubscript{Text: join(texts)}
}

// Superscript formats the given text as superscript (textSuperscript).
func Superscript(texts ...tg.RichTextClass) *tg.TextSuperscript {
	return &tg.TextSuperscript{Text: join(texts)}
}

// Marked formats the given text as marked / highlighted (textMarked).
func Marked(texts ...tg.RichTextClass) *tg.TextMarked {
	return &tg.TextMarked{Text: join(texts)}
}

// Spoiler formats the given text as a spoiler (textSpoiler).
func Spoiler(texts ...tg.RichTextClass) *tg.TextSpoiler {
	return &tg.TextSpoiler{Text: join(texts)}
}

// Mention formats the given text as a username mention (textMention).
func Mention(texts ...tg.RichTextClass) *tg.TextMention {
	return &tg.TextMention{Text: join(texts)}
}

// Hashtag formats the given text as a hashtag (textHashtag).
func Hashtag(texts ...tg.RichTextClass) *tg.TextHashtag {
	return &tg.TextHashtag{Text: join(texts)}
}

// BotCommand formats the given text as a bot command (textBotCommand).
func BotCommand(texts ...tg.RichTextClass) *tg.TextBotCommand {
	return &tg.TextBotCommand{Text: join(texts)}
}

// Cashtag formats the given text as a cashtag (textCashtag).
func Cashtag(texts ...tg.RichTextClass) *tg.TextCashtag {
	return &tg.TextCashtag{Text: join(texts)}
}

// BankCard formats the given text as a bank card number (textBankCard).
func BankCard(texts ...tg.RichTextClass) *tg.TextBankCard {
	return &tg.TextBankCard{Text: join(texts)}
}

// AutoURL formats the given text as an automatically detected URL (textAutoUrl).
func AutoURL(texts ...tg.RichTextClass) *tg.TextAutoURL {
	return &tg.TextAutoURL{Text: join(texts)}
}

// AutoEmail formats the given text as an automatically detected email address
// (textAutoEmail).
func AutoEmail(texts ...tg.RichTextClass) *tg.TextAutoEmail {
	return &tg.TextAutoEmail{Text: join(texts)}
}

// AutoPhone formats the given text as an automatically detected phone number
// (textAutoPhone).
func AutoPhone(texts ...tg.RichTextClass) *tg.TextAutoPhone {
	return &tg.TextAutoPhone{Text: join(texts)}
}

// URL formats the given text as a link to url (textUrl).
//
// webpageID may be zero if the linked webpage is not previewed.
func URL(text tg.RichTextClass, url string, webpageID int64) *tg.TextURL {
	return &tg.TextURL{Text: text, URL: url, WebpageID: webpageID}
}

// Email formats the given text as a link to an email address (textEmail).
func Email(text tg.RichTextClass, email string) *tg.TextEmail {
	return &tg.TextEmail{Text: text, Email: email}
}

// Phone formats the given text as a link to a phone number (textPhone).
func Phone(text tg.RichTextClass, phone string) *tg.TextPhone {
	return &tg.TextPhone{Text: text, Phone: phone}
}

// Anchor marks the given text as the target of an in-document anchor with the
// given name (textAnchor). See [AnchorLink] to link to it.
func Anchor(text tg.RichTextClass, name string) *tg.TextAnchor {
	return &tg.TextAnchor{Text: text, Name: name}
}

// AnchorLink links the given text to an in-document [Anchor] with the given
// name, rendered as a relative URL.
func AnchorLink(text tg.RichTextClass, name string) *tg.TextURL {
	return &tg.TextURL{Text: text, URL: "#" + name}
}

// Math returns an inline mathematical expression with the given LaTeX source
// (textMath).
func Math(source string) *tg.TextMath {
	return &tg.TextMath{Source: source}
}

// Image returns an inline image referencing a document by ID with the given
// size in pixels (textImage).
func Image(documentID int64, w, h int) *tg.TextImage {
	return &tg.TextImage{DocumentID: documentID, W: w, H: h}
}

// CustomEmoji returns an inline custom emoji referencing a document by ID, with
// the given fallback alt text (textCustomEmoji).
func CustomEmoji(documentID int64, alt string) *tg.TextCustomEmoji {
	return &tg.TextCustomEmoji{DocumentID: documentID, Alt: alt}
}

// MentionName formats the given text as a mention of a user by ID
// (textMentionName).
func MentionName(text tg.RichTextClass, userID int64) *tg.TextMentionName {
	return &tg.TextMentionName{Text: text, UserID: userID}
}

// DateFlags selects which components of a [Date] are rendered.
type DateFlags struct {
	// Relative renders the date relative to now ("5 minutes ago").
	Relative bool
	// ShortTime renders a short time.
	ShortTime bool
	// LongTime renders a long time.
	LongTime bool
	// ShortDate renders a short date.
	ShortDate bool
	// LongDate renders a long date.
	LongDate bool
	// DayOfWeek renders the day of the week.
	DayOfWeek bool
}

// Date formats the given text as the Unix timestamp date, rendered according to
// flags (textDate).
func Date(text tg.RichTextClass, date int, flags DateFlags) *tg.TextDate {
	return &tg.TextDate{
		Relative:  flags.Relative,
		ShortTime: flags.ShortTime,
		LongTime:  flags.LongTime,
		ShortDate: flags.ShortDate,
		LongDate:  flags.LongDate,
		DayOfWeek: flags.DayOfWeek,
		Text:      text,
		Date:      date,
	}
}
