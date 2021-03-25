package inline

import (
	"context"

	"github.com/gotd/td/tg"
)

// PhotoResultBuilder is photo result option builder.
type PhotoResultBuilder struct {
	result *tg.InputBotInlineResultPhoto
	msg    MessageOption
}

func (b *PhotoResultBuilder) apply(ctx context.Context, r *resultPageBuilder) error {
	m, err := b.msg.apply()
	if err != nil {
		return err
	}

	var t tg.InputBotInlineResultPhoto
	t.FillFrom(b.result)
	if t.ID == "" {
		t.ID, err = r.generateID()
		if err != nil {
			return err
		}
	}

	t.SendMessage = m
	r.results = append(r.results, &t)
	return nil
}

// ID sets ID of result.
// Should not be empty, so if id is not provided, random will be used.
func (b *PhotoResultBuilder) ID(id string) *PhotoResultBuilder {
	b.result.ID = id
	return b
}

// Photo creates game result option builder.
func Photo(photo tg.InputPhotoClass, msg MessageOption) *PhotoResultBuilder {
	return &PhotoResultBuilder{
		result: &tg.InputBotInlineResultPhoto{
			Type:  PhotoType,
			Photo: photo,
		},
		msg: msg,
	}
}
