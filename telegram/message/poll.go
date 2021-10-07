package message

import (
	"context"
	"strconv"
	"time"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/telegram/message/entity"
	"github.com/nnqq/td/telegram/message/styling"
	"github.com/nnqq/td/tg"
)

type pollAnswerBuilder struct {
	input *tg.InputMediaPoll
}

// PollAnswerOption is an option to create poll answer.
type PollAnswerOption func(p *pollAnswerBuilder)

// RawPollAnswer creates new raw poll answer option.
func RawPollAnswer(poll tg.PollAnswer) PollAnswerOption {
	return func(p *pollAnswerBuilder) {
		p.input.Poll.Answers = append(p.input.Poll.Answers, poll)
	}
}

// PollAnswer creates new plain poll answer option.
func PollAnswer(text string) PollAnswerOption {
	return func(p *pollAnswerBuilder) {
		i := len(p.input.Poll.Answers)
		p.input.Poll.Answers = append(p.input.Poll.Answers, tg.PollAnswer{
			Text:   text,
			Option: []byte(text + strconv.Itoa(i)),
		})
	}
}

// CorrectPollAnswer creates new correct poll answer option.
func CorrectPollAnswer(text string) PollAnswerOption {
	return func(p *pollAnswerBuilder) {
		p.input.Poll.Quiz = true
		i := len(p.input.Poll.Answers)
		option := []byte(text + strconv.Itoa(i))
		p.input.Poll.Answers = append(p.input.Poll.Answers, tg.PollAnswer{
			Text:   text,
			Option: option,
		})
		p.input.CorrectAnswers = append(p.input.CorrectAnswers, option)
	}
}

// PollBuilder is a Poll media option.
type PollBuilder struct {
	input   tg.InputMediaPoll
	answers []PollAnswerOption
	opts    []StyledTextOption
}

// PollID return poll ID. If poll was not sent, will be zero.
// It useful to close polls.
func (p *PollBuilder) PollID() int64 {
	return p.input.Poll.ID
}

// Close sets flag that the poll is closed and doesn't accept any more answers.
func (p *PollBuilder) Close() *PollBuilder {
	p.input.Poll.Closed = true
	return p
}

// PublicVoters sets flag that votes are publicly visible to all users (non-anonymous poll).
func (p *PollBuilder) PublicVoters(publicVoters bool) *PollBuilder {
	p.input.Poll.PublicVoters = publicVoters
	return p
}

// MultipleChoice sets flag that multiple options can be chosen as answer.
func (p *PollBuilder) MultipleChoice(multipleChoice bool) *PollBuilder {
	p.input.Poll.MultipleChoice = multipleChoice
	return p
}

// CloseDate sets point in time (Unix timestamp) when the poll will be automatically closed.
// Must be at least 5 and no more than 600 seconds in the future.
func (p *PollBuilder) CloseDate(d time.Time) *PollBuilder {
	return p.CloseDateTS(int(d.Unix()))
}

// CloseDateTS sets point in time (Unix timestamp) when the poll will be automatically closed.
// Must be at least 5 and no more than 600 seconds in the future.
func (p *PollBuilder) CloseDateTS(ts int) *PollBuilder {
	p.input.Poll.ClosePeriod = 0
	p.input.Poll.CloseDate = ts
	return p
}

// ClosePeriod sets amount of time in seconds the poll will be active after creation, 5-600 seconds.
func (p *PollBuilder) ClosePeriod(d time.Duration) *PollBuilder {
	return p.ClosePeriodSeconds(int(d.Seconds()))
}

// ClosePeriodSeconds sets amount of time in seconds the poll will be active after creation, 5-600.
func (p *PollBuilder) ClosePeriodSeconds(s int) *PollBuilder {
	p.input.Poll.ClosePeriod = s
	p.input.Poll.CloseDate = 0
	return p
}

// Explanation sets explanation message.
func (p *PollBuilder) Explanation(msg string) *PollBuilder {
	p.input.Solution = msg
	p.input.SolutionEntities = nil
	return p
}

// StyledExplanation sets styled explanation message.
func (p *PollBuilder) StyledExplanation(texts ...StyledTextOption) *PollBuilder {
	p.opts = texts
	return p
}

func (p *PollBuilder) apply(ctx context.Context, b *multiMediaBuilder) error {
	if p.input.Poll.ID == 0 {
		id, err := crypto.RandInt64(b.sender.rand)
		if err != nil {
			return xerrors.Errorf("generate id: %w", err)
		}

		p.input.Poll.ID = id
	}

	if len(p.opts) > 0 {
		tb := entity.Builder{}
		if err := styling.Perform(&tb, p.opts...); err != nil {
			return err
		}
		p.input.Solution, p.input.SolutionEntities = tb.Complete()
	}

	pb := pollAnswerBuilder{input: &p.input}
	for _, opt := range p.answers {
		opt(&pb)
	}

	return Media(&p.input).apply(ctx, b)
}

// Poll adds poll attachment.
func Poll(question string, a, b PollAnswerOption, answers ...PollAnswerOption) *PollBuilder {
	return &PollBuilder{
		input: tg.InputMediaPoll{
			Poll: tg.Poll{
				Question: question,
			},
		},
		answers: append([]PollAnswerOption{a, b}, answers...),
	}
}

// PollVote votes in a poll.
func (b *RequestBuilder) PollVote(
	ctx context.Context, msgID int,
	answer []byte, answers ...[]byte,
) (tg.UpdatesClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, xerrors.Errorf("peer: %w", err)
	}

	upd, err := b.sender.sendVote(ctx, &tg.MessagesSendVoteRequest{
		Peer:    p,
		MsgID:   msgID,
		Options: append([][]byte{answer}, answers...),
	})
	if err != nil {
		return nil, xerrors.Errorf("start bot: %w", err)
	}

	return upd, nil
}
