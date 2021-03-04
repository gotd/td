package message

import (
	"context"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestPoll(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	id := int64(0)
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMediaRequest)
		mock.True(ok)
		mock.Equal(&tg.InputPeerSelf{}, req.Peer)

		m, ok := req.Media.(*tg.InputMediaPoll)
		mock.True(ok)
		id = m.Poll.ID
		mock.Len(m.Poll.Answers, 3)
		mock.Len(m.CorrectAnswers, 1)
		mock.Equal(m.Poll.Answers[0].Option, m.CorrectAnswers[0])
	}).ThenResult(&tg.Updates{})
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMediaRequest)
		mock.True(ok)
		mock.Equal(&tg.InputPeerSelf{}, req.Peer)

		m, ok := req.Media.(*tg.InputMediaPoll)
		mock.True(ok)
		mock.Equal(id, m.Poll.ID)
		mock.True(m.Poll.Closed)
	}).ThenResult(&tg.Updates{})

	poll := Poll("Nu che tam s den'gami?",
		CorrectPollAnswer("A?"),
		PollAnswer("Che?"),
		PollAnswer("Kuda?"),
	).PublicVoters(true).
		StyledExplanation(
			Plain("See"), TextURL("https://youtu.be/PYzX7SDKhd0.", "https://youtu.be/PYzX7SDKhd0"),
		)

	mock.NoError(sender.Self().Media(ctx, poll))
	mock.NoError(sender.Self().Media(ctx, poll.Close()))
}

func TestRequestBuilder_PollVote(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectCall(&tg.MessagesSendVoteRequest{
		Peer:    &tg.InputPeerSelf{},
		MsgID:   10,
		Options: [][]byte{[]byte("abc")},
	}).ThenResult(&tg.Updates{})
	mock.NoError(sender.Self().PollVote(ctx, 10, []byte("abc")))

	mock.ExpectCall(&tg.MessagesSendVoteRequest{
		Peer:    &tg.InputPeerSelf{},
		MsgID:   10,
		Options: [][]byte{[]byte("abc")},
	}).ThenRPCErr(testRPCError())
	mock.Error(sender.Self().PollVote(ctx, 10, []byte("abc")))
}
