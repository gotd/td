package message

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/telegram/internal/rpcmock"
	"github.com/gotd/td/tg"
)

func expectSendSetTyping(action tg.SendMessageActionClass, mock *rpcmock.Mock, threadID int) {
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSetTypingRequest)
		mock.True(ok)
		mock.Equal(&tg.InputPeerSelf{}, req.Peer)
		mock.Equal(action, req.Action)
		mock.Equal(threadID, req.TopMsgID)
	}).ThenTrue()
}

func TestRequestBuilder_TypingAction(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	id64, err := crypto.RandInt64(rand.Reader)
	mock.NoError(err)
	id := int(id64)

	act := sender.Self().TypingAction().ThreadMsg(&tg.Message{ID: id})

	expectSendSetTyping(&tg.SendMessageTypingAction{}, mock, id)
	mock.NoError(act.Typing(ctx))
	expectSendSetTyping(&tg.SendMessageCancelAction{}, mock, id)
	mock.NoError(act.Cancel(ctx))
	expectSendSetTyping(&tg.SendMessageRecordVideoAction{}, mock, id)
	mock.NoError(act.RecordVideo(ctx))
	expectSendSetTyping(&tg.SendMessageUploadVideoAction{Progress: 10}, mock, id)
	mock.NoError(act.UploadVideo(ctx, 10))
	expectSendSetTyping(&tg.SendMessageRecordAudioAction{}, mock, id)
	mock.NoError(act.RecordAudio(ctx))
	expectSendSetTyping(&tg.SendMessageUploadAudioAction{Progress: 10}, mock, id)
	mock.NoError(act.UploadAudio(ctx, 10))
	expectSendSetTyping(&tg.SendMessageUploadPhotoAction{Progress: 10}, mock, id)
	mock.NoError(act.UploadPhoto(ctx, 10))
	expectSendSetTyping(&tg.SendMessageUploadDocumentAction{Progress: 10}, mock, id)
	mock.NoError(act.UploadDocument(ctx, 10))
	expectSendSetTyping(&tg.SendMessageGeoLocationAction{}, mock, id)
	mock.NoError(act.GeoLocation(ctx))
	expectSendSetTyping(&tg.SendMessageChooseContactAction{}, mock, id)
	mock.NoError(act.ChooseContact(ctx))
	expectSendSetTyping(&tg.SendMessageGamePlayAction{}, mock, id)
	mock.NoError(act.GamePlay(ctx))
	expectSendSetTyping(&tg.SendMessageRecordRoundAction{}, mock, id)
	mock.NoError(act.RecordRound(ctx))
	expectSendSetTyping(&tg.SendMessageUploadRoundAction{Progress: 10}, mock, id)
	mock.NoError(act.UploadRound(ctx, 10))
	expectSendSetTyping(&tg.SpeakingInGroupCallAction{}, mock, id)
	mock.NoError(act.SpeakingInGroupCall(ctx))
	expectSendSetTyping(&tg.SendMessageHistoryImportAction{Progress: 10}, mock, id)
	mock.NoError(act.HistoryImport(ctx, 10))
}
