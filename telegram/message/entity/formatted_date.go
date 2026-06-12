package entity

import "github.com/go-faster/errors"

// FormattedDateFormat parses a Telegram date format string, as used by the
// HTML <tg-time format="..."> attribute and the markdown tg://time?format=...
// URL, and returns a FormattedDate Formatter rendering the given Unix timestamp.
//
// The format string is a set of flag characters:
//
//	r, R  relative time ("5 minutes ago"); must be the only character
//	t     short time
//	T     long time
//	d     short date
//	D     long date
//	w, W  day of week
//
// An empty format string sets no flags. An unrecognized character returns an
// error.
//
// See https://core.telegram.org/constructor/messageEntityFormattedDate.
func FormattedDateFormat(format string, date int) (Formatter, error) {
	var (
		relative, shortTime, longTime  bool
		shortDate, longDate, dayOfWeek bool
	)
	// "r"/"R" means relative time and must be the only character, matching
	// tdlib's get_date_flags.
	if format == "r" || format == "R" {
		relative = true
	} else {
		for _, c := range format {
			switch c {
			case 't':
				shortTime = true
			case 'T':
				longTime = true
			case 'd':
				shortDate = true
			case 'D':
				longDate = true
			case 'w', 'W':
				dayOfWeek = true
			default:
				return nil, errors.Errorf("invalid date format character %q", c)
			}
		}
	}
	return FormattedDate(relative, shortTime, longTime, shortDate, longDate, dayOfWeek, date), nil
}
