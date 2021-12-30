package styling

import (
	"github.com/gotd/td/telegram/message/entity"
)

// StyledTextOption is an option for styling text.
type StyledTextOption struct {
	size    int
	perform func(b *textBuilder) error
}

// Zero returns true if option is zero value.
func (s StyledTextOption) Zero() bool {
	return s.perform == nil
}

func styledTextOption(s string, perform func(b *textBuilder) error) StyledTextOption {
	return StyledTextOption{
		perform: perform,
		size:    len(s),
	}
}

// Plain formats text without any entities.
func Plain(s string) StyledTextOption {
	return styledTextOption(s, func(b *textBuilder) error {
		b.Plain(s)
		return nil
	})
}

// Custom formats text using given callback.
func Custom(cb func(eb *entity.Builder) error) StyledTextOption {
	return StyledTextOption{
		size: 0,
		perform: func(b *textBuilder) error {
			return cb(b.Builder)
		},
	}
}

//go:generate go run github.com/gotd/td/telegram/message/internal/mkentity -template styling -output options.gen.go
