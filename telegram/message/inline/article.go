package inline

import (
	"github.com/nnqq/td/tg"
)

// ArticleResultBuilder is article result option builder.
type ArticleResultBuilder struct {
	result *tg.InputBotInlineResult
	msg    MessageOption
}

func (b *ArticleResultBuilder) apply(r *resultPageBuilder) error {
	m, err := b.msg.apply()
	if err != nil {
		return err
	}

	t := tg.InputBotInlineResult{
		ID:          b.result.ID,
		Type:        b.result.Type,
		Title:       b.result.Title,
		Description: b.result.Description,
		URL:         b.result.URL,
		Thumb:       b.result.Thumb,
		Content:     b.result.Content,
	}
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

// Title sets Result description.
func (b *ArticleResultBuilder) Title(title string) *ArticleResultBuilder {
	b.result.SetTitle(title)
	return b
}

// Description sets Result description.
func (b *ArticleResultBuilder) Description(description string) *ArticleResultBuilder {
	b.result.SetDescription(description)
	return b
}

// URL sets URL of result.
func (b *ArticleResultBuilder) URL(url string) *ArticleResultBuilder {
	b.result.SetURL(url)
	return b
}

// Thumb sets Thumbnail for result.
func (b *ArticleResultBuilder) Thumb(thumb tg.InputWebDocument) *ArticleResultBuilder {
	b.result.SetThumb(thumb)
	return b
}

// Content sets Result contents.
func (b *ArticleResultBuilder) Content(content tg.InputWebDocument) *ArticleResultBuilder {
	b.result.SetContent(content)
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
