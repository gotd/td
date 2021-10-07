package message

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgmock"
)

func expectSendSetTyping(t *testing.T, action tg.SendMessageActionClass, mock *tgmock.Mock, threadID int) {
	t.Helper()
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSetTypingRequest)
		require.True(t, ok)
		require.Equal(t, &tg.InputPeerSelf{}, req.Peer)
		require.Equal(t, action, req.Action)
		require.Equal(t, threadID, req.TopMsgID)
	}).ThenTrue()
}

func TestRequestBuilder_TypingAction(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	id64, err := crypto.RandInt64(rand.Reader)
	require.NoError(t, err)
	id := int(id64)

	act := sender.Self().TypingAction().ThreadMsg(&tg.Message{ID: id})

	expectSendSetTyping(t, &tg.SendMessageTypingAction{}, mock, id)
	require.NoError(t, act.Typing(ctx))
	expectSendSetTyping(t, &tg.SendMessageCancelAction{}, mock, id)
	require.NoError(t, act.Cancel(ctx))
	expectSendSetTyping(t, &tg.SendMessageRecordVideoAction{}, mock, id)
	require.NoError(t, act.RecordVideo(ctx))
	expectSendSetTyping(t, &tg.SendMessageUploadVideoAction{Progress: 10}, mock, id)
	require.NoError(t, act.UploadVideo(ctx, 10))
	expectSendSetTyping(t, &tg.SendMessageRecordAudioAction{}, mock, id)
	require.NoError(t, act.RecordAudio(ctx))
	expectSendSetTyping(t, &tg.SendMessageUploadAudioAction{Progress: 10}, mock, id)
	require.NoError(t, act.UploadAudio(ctx, 10))
	expectSendSetTyping(t, &tg.SendMessageUploadPhotoAction{Progress: 10}, mock, id)
	require.NoError(t, act.UploadPhoto(ctx, 10))
	expectSendSetTyping(t, &tg.SendMessageUploadDocumentAction{Progress: 10}, mock, id)
	require.NoError(t, act.UploadDocument(ctx, 10))
	expectSendSetTyping(t, &tg.SendMessageGeoLocationAction{}, mock, id)
	require.NoError(t, act.GeoLocation(ctx))
	expectSendSetTyping(t, &tg.SendMessageChooseContactAction{}, mock, id)
	require.NoError(t, act.ChooseContact(ctx))
	expectSendSetTyping(t, &tg.SendMessageGamePlayAction{}, mock, id)
	require.NoError(t, act.GamePlay(ctx))
	expectSendSetTyping(t, &tg.SendMessageRecordRoundAction{}, mock, id)
	require.NoError(t, act.RecordRound(ctx))
	expectSendSetTyping(t, &tg.SendMessageUploadRoundAction{Progress: 10}, mock, id)
	require.NoError(t, act.UploadRound(ctx, 10))
	expectSendSetTyping(t, &tg.SpeakingInGroupCallAction{}, mock, id)
	require.NoError(t, act.SpeakingInGroupCall(ctx))
	expectSendSetTyping(t, &tg.SendMessageHistoryImportAction{Progress: 10}, mock, id)
	require.NoError(t, act.HistoryImport(ctx, 10))
}
