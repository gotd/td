package inline

import (
	"io"
	"strconv"

	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/tg"
)

type resultPageBuilder struct {
	results []tg.InputBotInlineResultClass
	random  io.Reader
}

func (r *resultPageBuilder) generateID() (string, error) {
	n, err := crypto.RandInt64(r.random)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(n, 10), nil
}

// ResultOption is an option of inline result.
type ResultOption interface {
	apply(r *resultPageBuilder) error
}

// MessageOption is an option of inline result message.
type MessageOption interface {
	apply() (tg.InputBotInlineMessageClass, error)
}

type messageOptionFunc func() (tg.InputBotInlineMessageClass, error)

func (m messageOptionFunc) apply() (tg.InputBotInlineMessageClass, error) {
	return m()
}

// ResultMessage creates new MessageOption from given message object.
func ResultMessage(r tg.InputBotInlineMessageClass) MessageOption {
	return messageOptionFunc(func() (tg.InputBotInlineMessageClass, error) {
		return r, nil
	})
}
