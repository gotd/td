package styling

import (
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/message/entity"
)

type textBuilder struct {
	*entity.Builder
}

func (b *textBuilder) Perform(text StyledTextOption, texts ...StyledTextOption) error {
	b.GrowEntities(len(texts) + 1)
	length := text.size
	for i := range texts {
		length += texts[i].size
	}
	b.GrowText(length)

	if err := text.perform(b); err != nil {
		return xerrors.Errorf("perform first styling option: %w", err)
	}
	for idx, opt := range texts {
		if err := opt.perform(b); err != nil {
			return xerrors.Errorf("perform %d styling option: %w", idx, err)
		}
	}

	return nil
}

// Perform performs all options to the given builder.
func Perform(builder *entity.Builder, text StyledTextOption, texts ...StyledTextOption) error {
	tb := &textBuilder{Builder: builder}
	return tb.Perform(text, texts...)
}
