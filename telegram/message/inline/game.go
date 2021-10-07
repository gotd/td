package inline

import (
	"github.com/nnqq/td/tg"
)

// GameResultBuilder is game result option builder.
type GameResultBuilder struct {
	result *tg.InputBotInlineResultGame
	msg    MessageOption
}

func (b *GameResultBuilder) apply(r *resultPageBuilder) error {
	m, err := b.msg.apply()
	if err != nil {
		return err
	}

	t := tg.InputBotInlineResultGame{
		ID:        b.result.ID,
		ShortName: b.result.ShortName,
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
func (b *GameResultBuilder) ID(id string) *GameResultBuilder {
	b.result.ID = id
	return b
}

// Game creates game result option builder.
func Game(shortName string, msg MessageOption) *GameResultBuilder {
	return &GameResultBuilder{
		result: &tg.InputBotInlineResultGame{
			ShortName: shortName,
		},
		msg: msg,
	}
}
