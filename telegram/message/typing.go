package message

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/tg"
)

// TypingActionBuilder is a helper to create and send typing actions.
//
// See https://core.telegram.org/method/messages.setTyping.
//
// See https://core.telegram.org/type/SendMessageAction.
type TypingActionBuilder struct {
	sender   *Sender
	peer     peerPromise
	topMsgID int
}

// ThreadID sets thread ID to send.
func (b *TypingActionBuilder) ThreadID(id int) *TypingActionBuilder {
	b.topMsgID = id
	return b
}

// ThreadMsg sets message's ID as thread ID to send.
func (b *TypingActionBuilder) ThreadMsg(msg tg.MessageClass) *TypingActionBuilder {
	return b.ThreadID(msg.GetID())
}

func (b *TypingActionBuilder) send(ctx context.Context, action tg.SendMessageActionClass) error {
	p, err := b.peer(ctx)
	if err != nil {
		return xerrors.Errorf("peer: %w", err)
	}

	if err := b.sender.setTyping(ctx, &tg.MessagesSetTypingRequest{
		Peer:     p,
		TopMsgID: b.topMsgID,
		Action:   action,
	}); err != nil {
		return xerrors.Errorf("set typing: %w", err)
	}

	return nil
}

// Typing sends SendMessageTypingAction.
func (b *TypingActionBuilder) Typing(ctx context.Context) error {
	return b.send(ctx, &tg.SendMessageTypingAction{})
}

// Cancel sends SendMessageCancelAction.
func (b *TypingActionBuilder) Cancel(ctx context.Context) error {
	return b.send(ctx, &tg.SendMessageCancelAction{})
}

// RecordVideo sends SendMessageRecordVideoAction.
func (b *TypingActionBuilder) RecordVideo(ctx context.Context) error {
	return b.send(ctx, &tg.SendMessageRecordVideoAction{})
}

// UploadVideo sends SendMessageUploadVideoAction.
func (b *TypingActionBuilder) UploadVideo(ctx context.Context, progress int) error {
	return b.send(ctx, &tg.SendMessageUploadVideoAction{Progress: progress})
}

// RecordAudio sends SendMessageRecordAudioAction.
func (b *TypingActionBuilder) RecordAudio(ctx context.Context) error {
	return b.send(ctx, &tg.SendMessageRecordAudioAction{})
}

// UploadAudio sends SendMessageUploadAudioAction.
func (b *TypingActionBuilder) UploadAudio(ctx context.Context, progress int) error {
	return b.send(ctx, &tg.SendMessageUploadAudioAction{Progress: progress})
}

// UploadPhoto sends SendMessageUploadPhotoAction.
func (b *TypingActionBuilder) UploadPhoto(ctx context.Context, progress int) error {
	return b.send(ctx, &tg.SendMessageUploadPhotoAction{Progress: progress})
}

// UploadDocument sends SendMessageUploadDocumentAction.
func (b *TypingActionBuilder) UploadDocument(ctx context.Context, progress int) error {
	return b.send(ctx, &tg.SendMessageUploadDocumentAction{Progress: progress})
}

// GeoLocation sends SendMessageGeoLocationAction.
func (b *TypingActionBuilder) GeoLocation(ctx context.Context) error {
	return b.send(ctx, &tg.SendMessageGeoLocationAction{})
}

// ChooseContact sends SendMessageChooseContactAction.
func (b *TypingActionBuilder) ChooseContact(ctx context.Context) error {
	return b.send(ctx, &tg.SendMessageChooseContactAction{})
}

// GamePlay sends SendMessageGamePlayAction.
func (b *TypingActionBuilder) GamePlay(ctx context.Context) error {
	return b.send(ctx, &tg.SendMessageGamePlayAction{})
}

// RecordRound sends SendMessageRecordRoundAction.
func (b *TypingActionBuilder) RecordRound(ctx context.Context) error {
	return b.send(ctx, &tg.SendMessageRecordRoundAction{})
}

// UploadRound sends SendMessageUploadRoundAction.
func (b *TypingActionBuilder) UploadRound(ctx context.Context, progress int) error {
	return b.send(ctx, &tg.SendMessageUploadRoundAction{Progress: progress})
}

// SpeakingInGroupCall sends SpeakingInGroupCallAction.
func (b *TypingActionBuilder) SpeakingInGroupCall(ctx context.Context) error {
	return b.send(ctx, &tg.SpeakingInGroupCallAction{})
}

// HistoryImport sends SendMessageHistoryImportAction.
func (b *TypingActionBuilder) HistoryImport(ctx context.Context, progress int) error {
	return b.send(ctx, &tg.SendMessageHistoryImportAction{Progress: progress})
}

// TypingAction creates TypingActionBuilder.
func (b *RequestBuilder) TypingAction() *TypingActionBuilder {
	return &TypingActionBuilder{
		sender: b.sender,
		peer:   b.peer,
	}
}
