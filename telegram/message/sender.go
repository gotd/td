package message

import (
	"context"
	"crypto/rand"
	"io"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
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
		rand:     rand.Reader,
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
func (s *Sender) sendMessage(ctx context.Context, req *tg.MessagesSendMessageRequest) error {
	if req.RandomID == 0 {
		id, err := crypto.RandInt64(s.rand)
		if err != nil {
			return xerrors.Errorf("generate random_id: %w", err)
		}
		req.RandomID = id
	}

	_, err := s.raw.MessagesSendMessage(ctx, req)
	return err
}

// sendMedia sends message with single media to peer.
func (s *Sender) sendMedia(ctx context.Context, req *tg.MessagesSendMediaRequest) error {
	if req.RandomID == 0 {
		id, err := crypto.RandInt64(s.rand)
		if err != nil {
			return xerrors.Errorf("generate random_id: %w", err)
		}
		req.RandomID = id
	}

	_, err := s.raw.MessagesSendMedia(ctx, req)
	return err
}

// sendMultiMedia sends message with multiple media to peer.
func (s *Sender) sendMultiMedia(ctx context.Context, req *tg.MessagesSendMultiMediaRequest) error {
	for i := range req.MultiMedia {
		id, err := crypto.RandInt64(s.rand)
		if err != nil {
			return xerrors.Errorf("generate random_id: %w", err)
		}
		req.MultiMedia[i].RandomID = id
	}

	_, err := s.raw.MessagesSendMultiMedia(ctx, req)
	return err
}

// forwardMessages forwards message to peer.
func (s *Sender) forwardMessages(ctx context.Context, req *tg.MessagesForwardMessagesRequest) error {
	req.RandomID = make([]int64, len(req.ID))
	for i := range req.RandomID {
		id, err := crypto.RandInt64(s.rand)
		if err != nil {
			return xerrors.Errorf("generate random_id: %w", err)
		}
		req.RandomID[i] = id
	}

	_, err := s.raw.MessagesForwardMessages(ctx, req)
	return err
}

// startBot starts a conversation with a bot using a deep linking parameter.
func (s *Sender) startBot(ctx context.Context, req *tg.MessagesStartBotRequest) error {
	if req.RandomID == 0 {
		id, err := crypto.RandInt64(s.rand)
		if err != nil {
			return xerrors.Errorf("generate random_id: %w", err)
		}
		req.RandomID = id
	}

	_, err := s.raw.MessagesStartBot(ctx, req)
	return err
}

// sendInlineBotResult sends inline query result message to peer.
func (s *Sender) sendInlineBotResult(ctx context.Context, req *tg.MessagesSendInlineBotResultRequest) error {
	if req.RandomID == 0 {
		id, err := crypto.RandInt64(s.rand)
		if err != nil {
			return xerrors.Errorf("generate random_id: %w", err)
		}
		req.RandomID = id
	}

	_, err := s.raw.MessagesSendInlineBotResult(ctx, req)
	return err
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
func (s *Sender) sendVote(ctx context.Context, req *tg.MessagesSendVoteRequest) error {
	_, err := s.raw.MessagesSendVote(ctx, req)
	return err
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
) error {
	if req.RandomID == 0 {
		id, err := crypto.RandInt64(s.rand)
		if err != nil {
			return xerrors.Errorf("generate random_id: %w", err)
		}
		req.RandomID = id
	}

	_, err := s.raw.MessagesSendScreenshotNotification(ctx, req)
	return err
}
