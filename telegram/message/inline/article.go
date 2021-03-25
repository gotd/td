package inline

import (
	"context"

	"github.com/gotd/td/tg"
)

// ArticleResultBuilder is article result option builder.
type ArticleResultBuilder struct {
	result *tg.InputBotInlineResult
	msg    MessageOption
}

func (b *ArticleResultBuilder) apply(ctx context.Context, r *resultPageBuilder) error {
	m, err := b.msg.apply()
	if err != nil {
		return err
	}

	var t tg.InputBotInlineResult
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
func (b *ArticleResultBuilder) ID(id string) *ArticleResultBuilder {
	b.result.ID = id
	return b
}

// Type sets Result type (see bot API docs¹).
func (b *ArticleResultBuilder) Type(typ string) *ArticleResultBuilder {
	b.result.Type = typ
	return b
}

// Description sets Result description.
func (b *ArticleResultBuilder) Description(description string) *ArticleResultBuilder {
	b.result.Description = description
	return b
}

// URL sets URL of result.
func (b *ArticleResultBuilder) URL(url string) *ArticleResultBuilder {
	b.result.URL = url
	return b
}

// Thumb sets Thumbnail for result.
func (b *ArticleResultBuilder) Thumb(thumb tg.InputWebDocument) *ArticleResultBuilder {
	b.result.Thumb = thumb
	return b
}

// Content sets Result contents.
func (b *ArticleResultBuilder) Content(content tg.InputWebDocument) *ArticleResultBuilder {
	b.result.Content = content
	return b
}

// Article creates article result option builder.
func Article(title string, msg MessageOption) *ArticleResultBuilder {
	return &ArticleResultBuilder{
		result: &tg.InputBotInlineResult{
			Type:  ArticleType,
			Title: title,
		},
		msg: msg,
	}
}
