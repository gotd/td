package styling

import (
	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/message/entity"
)

type textBuilder struct {
	*entity.Builder
}

func (b *textBuilder) Perform(texts ...StyledTextOption) error {
	b.GrowEntities(len(texts))
	var length int
	for i := range texts {
		length += texts[i].size
	}
	b.GrowText(length)

	for idx, opt := range texts {
		if err := opt.perform(b); err != nil {
			return errors.Wrapf(err, "perform %d styling option", idx+2)
		}
	}

	return nil
}

// Perform performs all options to the given builder.
func Perform(builder *entity.Builder, texts ...StyledTextOption) error {
	tb := &textBuilder{Builder: builder}
	return tb.Perform(texts...)
}
