package message

import (
	"context"
	"crypto/rand"
	"io"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

// Sender is a message sending helper.
type Sender struct {
	raw  *tg.Client
	rand io.Reader

	uploader *uploader.Uploader
	resolver PeerResolver
}

// NewSender creates a new Sender.
func NewSender(raw *tg.Client) *Sender {
	return &Sender{
		raw:      raw,
		rand:     rand.Reader,
		uploader: uploader.NewUploader(raw),
		resolver: DefaultPeerResolver(raw),
	}
}

// WithUploader sets file uploader to use.
func (s *Sender) WithUploader(u *uploader.Uploader) *Sender {
	s.uploader = u
	return s
}

// WithResolver sets peer resolver to use.
func (s *Sender) WithResolver(resolver PeerResolver) *Sender {
	s.resolver = resolver
	return s
}

// WithRand sets random ID source.
func (s *Sender) WithRand(r io.Reader) *Sender {
	s.rand = r
	return s
}

// SendMessage sends message to peer.
func (s *Sender) SendMessage(ctx context.Context, req *tg.MessagesSendMessageRequest) error {
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

// SendMedia sends message with single media to peer.
func (s *Sender) SendMedia(ctx context.Context, req *tg.MessagesSendMediaRequest) error {
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

// SendMultiMedia sends message with multiple media to peer.
func (s *Sender) SendMultiMedia(ctx context.Context, req *tg.MessagesSendMultiMediaRequest) error {
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

// UploadMedia uploads file and associate it to a chat (without actually sending it to the chat).
func (s *Sender) UploadMedia(ctx context.Context, req *tg.MessagesUploadMediaRequest) (tg.MessageMediaClass, error) {
	return s.raw.MessagesUploadMedia(ctx, req)
}

// SendScreenshotNotification sends notification about screenshot to peer.
func (s *Sender) SendScreenshotNotification(
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
