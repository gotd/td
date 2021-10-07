package styling

import (
	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram/message/entity"
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
			return xerrors.Errorf("perform %d styling option: %w", idx+2, err)
		}
	}

	return nil
}

// Perform performs all options to the given builder.
func Perform(builder *entity.Builder, texts ...StyledTextOption) error {
	tb := &textBuilder{Builder: builder}
	return tb.Perform(texts...)
}
