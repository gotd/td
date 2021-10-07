package message

import (
	"context"
	"io"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/telegram/message/peer"
	"github.com/nnqq/td/telegram/uploader"
	"github.com/nnqq/td/tg"
)

// Sender is a message sending helper.
type Sender struct {
	raw  *tg.Client
	rand io.Reader

	uploader Uploader
	resolver peer.Resolver
}

// NewSender creates a new Sender.
func NewSender(raw *tg.Client) *Sender {
	return &Sender{
		raw:      raw,
		rand:     crypto.DefaultRand(),
		uploader: uploader.NewUploader(raw),
		resolver: peer.DefaultResolver(raw),
	}
}

// WithUploader sets file uploader to use.
func (s *Sender) WithUploader(u Uploader) *Sender {
	s.uploader = u
	return s
}

// WithResolver sets peer resolver to use.
func (s *Sender) WithResolver(resolver peer.Resolver) *Sender {
	s.resolver = resolver
	return s
}

// WithRand sets random ID source.
func (s *Sender) WithRand(r io.Reader) *Sender {
	s.rand = r
	return s
}

// ClearAllDrafts clears all drafts in all peers.
func (s *Sender) ClearAllDrafts(ctx context.Context) error {
	_, err := s.raw.MessagesClearAllDrafts(ctx)
	return err
}

// sendMessage sends message to peer.
func (s *Sender) sendMessage(ctx context.Context, req *tg.MessagesSendMessageRequest) (tg.UpdatesClass, error) {
	if req.RandomID == 0 {
		id, err := crypto.RandInt64(s.rand)
		if err != nil {
			return nil, xerrors.Errorf("generate random_id: %w", err)
		}
		req.RandomID = id
	}

	return s.raw.MessagesSendMessage(ctx, req)
}

// sendMedia sends message with single media to peer.
func (s *Sender) sendMedia(ctx context.Context, req *tg.MessagesSendMediaRequest) (tg.UpdatesClass, error) {
	if req.RandomID == 0 {
		id, err := crypto.RandInt64(s.rand)
		if err != nil {
			return nil, xerrors.Errorf("generate random_id: %w", err)
		}
		req.RandomID = id
	}

	return s.raw.MessagesSendMedia(ctx, req)
}

// sendMultiMedia sends message with multiple media to peer.
func (s *Sender) sendMultiMedia(ctx context.Context, req *tg.MessagesSendMultiMediaRequest) (tg.UpdatesClass, error) {
	for i := range req.MultiMedia {
		id, err := crypto.RandInt64(s.rand)
		if err != nil {
			return nil, xerrors.Errorf("generate random_id: %w", err)
		}
		req.MultiMedia[i].RandomID = id
	}

	return s.raw.MessagesSendMultiMedia(ctx, req)
}

// editMessage edits message.
func (s *Sender) editMessage(ctx context.Context, req *tg.MessagesEditMessageRequest) (tg.UpdatesClass, error) {
	return s.raw.MessagesEditMessage(ctx, req)
}

// forwardMessages forwards message to peer.
func (s *Sender) forwardMessages(ctx context.Context, req *tg.MessagesForwardMessagesRequest) (tg.UpdatesClass, error) {
	req.RandomID = make([]int64, len(req.ID))
	for i := range req.RandomID {
		id, err := crypto.RandInt64(s.rand)
		if err != nil {
			return nil, xerrors.Errorf("generate random_id: %w", err)
		}
		req.RandomID[i] = id
	}

	return s.raw.MessagesForwardMessages(ctx, req)
}

// startBot starts a conversation with a bot using a deep linking parameter.
func (s *Sender) startBot(ctx context.Context, req *tg.MessagesStartBotRequest) (tg.UpdatesClass, error) {
	if req.RandomID == 0 {
		id, err := crypto.RandInt64(s.rand)
		if err != nil {
			return nil, xerrors.Errorf("generate random_id: %w", err)
		}
		req.RandomID = id
	}

	return s.raw.MessagesStartBot(ctx, req)
}

// sendInlineBotResult sends inline query result message to peer.
func (s *Sender) sendInlineBotResult(
	ctx context.Context,
	req *tg.MessagesSendInlineBotResultRequest,
) (tg.UpdatesClass, error) {
	if req.RandomID == 0 {
		id, err := crypto.RandInt64(s.rand)
		if err != nil {
			return nil, xerrors.Errorf("generate random_id: %w", err)
		}
		req.RandomID = id
	}

	return s.raw.MessagesSendInlineBotResult(ctx, req)
}

// uploadMedia uploads file and associate it to a chat (without actually sending it to the chat).
func (s *Sender) uploadMedia(ctx context.Context, req *tg.MessagesUploadMediaRequest) (tg.MessageMediaClass, error) {
	return s.raw.MessagesUploadMedia(ctx, req)
}

// getDocumentByHash finds document by hash, MIME type and size.
func (s *Sender) getDocumentByHash(
	ctx context.Context,
	req *tg.MessagesGetDocumentByHashRequest,
) (tg.DocumentClass, error) {
	return s.raw.MessagesGetDocumentByHash(ctx, req)
}

// saveDraft saves a message draft associated to a chat.
func (s *Sender) saveDraft(ctx context.Context, req *tg.MessagesSaveDraftRequest) error {
	_, err := s.raw.MessagesSaveDraft(ctx, req)
	return err
}

// sendVote votes in a poll.
func (s *Sender) sendVote(ctx context.Context, req *tg.MessagesSendVoteRequest) (tg.UpdatesClass, error) {
	return s.raw.MessagesSendVote(ctx, req)
}

// setTyping sends a typing event to a conversation partner or group.
func (s *Sender) setTyping(ctx context.Context, req *tg.MessagesSetTypingRequest) error {
	_, err := s.raw.MessagesSetTyping(ctx, req)
	return err
}

// report reports a message in a chat for violation of Telegram's Terms of Service.
func (s *Sender) report(ctx context.Context, req *tg.MessagesReportRequest) (bool, error) {
	return s.raw.MessagesReport(ctx, req)
}

// reportSpam reports a new incoming chat for spam, if the peer settings of the chat allow us to do that.
func (s *Sender) reportSpam(ctx context.Context, p tg.InputPeerClass) (bool, error) {
	return s.raw.MessagesReportSpam(ctx, p)
}

// getPeerSettings returns peer settings.
func (s *Sender) getPeerSettings(ctx context.Context, p tg.InputPeerClass) (*tg.PeerSettings, error) {
	return s.raw.MessagesGetPeerSettings(ctx, p)
}

// sendScreenshotNotification sends notification about screenshot to peer.
func (s *Sender) sendScreenshotNotification(
	ctx context.Context,
	req *tg.MessagesSendScreenshotNotificationRequest,
) (tg.UpdatesClass, error) {
	if req.RandomID == 0 {
		id, err := crypto.RandInt64(s.rand)
		if err != nil {
			return nil, xerrors.Errorf("generate random_id: %w", err)
		}
		req.RandomID = id
	}

	return s.raw.MessagesSendScreenshotNotification(ctx, req)
}

// sendScheduledMessages sends scheduled messages using given ids.
func (s *Sender) sendScheduledMessages(
	ctx context.Context,
	req *tg.MessagesSendScheduledMessagesRequest,
) (tg.UpdatesClass, error) {
	return s.raw.MessagesSendScheduledMessages(ctx, req)
}

// deleteScheduledMessages deletes scheduled messages using given ids.
func (s *Sender) deleteScheduledMessages(
	ctx context.Context,
	req *tg.MessagesDeleteScheduledMessagesRequest,
) (tg.UpdatesClass, error) {
	return s.raw.MessagesDeleteScheduledMessages(ctx, req)
}

// getScheduledHistory gets scheduled messages history.
func (s *Sender) getScheduledHistory(
	ctx context.Context,
	req *tg.MessagesGetScheduledHistoryRequest,
) (tg.MessagesMessagesClass, error) {
	return s.raw.MessagesGetScheduledHistory(ctx, req)
}

// getScheduledMessages gets scheduled messages using given ids.
func (s *Sender) getScheduledMessages(
	ctx context.Context,
	req *tg.MessagesGetScheduledMessagesRequest,
) (tg.MessagesMessagesClass, error) {
	return s.raw.MessagesGetScheduledMessages(ctx, req)
}

// importChatInvite imports a chat invite and join a private chat/supergroup/channel.
func (s *Sender) importChatInvite(
	ctx context.Context,
	hash string,
) (tg.UpdatesClass, error) {
	return s.raw.MessagesImportChatInvite(ctx, hash)
}

// joinChannel joins a channel/supergroup.
func (s *Sender) joinChannel(
	ctx context.Context,
	input tg.InputChannelClass,
) (tg.UpdatesClass, error) {
	return s.raw.ChannelsJoinChannel(ctx, input)
}

// leaveChannel leaves a channel/supergroup.
func (s *Sender) leaveChannel(
	ctx context.Context,
	input tg.InputChannelClass,
) (tg.UpdatesClass, error) {
	return s.raw.ChannelsLeaveChannel(ctx, input)
}

// deleteChatUser delete user from chat.
func (s *Sender) deleteChatUser(
	ctx context.Context,
	req *tg.MessagesDeleteChatUserRequest,
) (tg.UpdatesClass, error) {
	return s.raw.MessagesDeleteChatUser(ctx, req)
}

// deleteChannelMessages deletes messages in channel.
func (s *Sender) deleteChannelMessages(
	ctx context.Context,
	req *tg.ChannelsDeleteMessagesRequest,
) (*tg.MessagesAffectedMessages, error) {
	return s.raw.ChannelsDeleteMessages(ctx, req)
}

// deleteMessages deletes messages in chat.
func (s *Sender) deleteMessages(
	ctx context.Context,
	req *tg.MessagesDeleteMessagesRequest,
) (*tg.MessagesAffectedMessages, error) {
	return s.raw.MessagesDeleteMessages(ctx, req)
}
