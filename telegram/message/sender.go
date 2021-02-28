package message

import (
	"context"
	"crypto/rand"
	"io"

	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/tg"
)

// Sender is a message sending helper.
type Sender struct {
	raw  *tg.Client
	rand io.Reader
}

// NewSender creates a new Sender.
func NewSender(raw *tg.Client) Sender {
	return Sender{
		raw:  raw,
		rand: rand.Reader,
	}
}

// SendMessage sends message to peer.
func (s Sender) SendMessage(ctx context.Context, req *tg.MessagesSendMessageRequest) error {
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

// SendMedia sends message to peer.
func (s Sender) SendMedia(ctx context.Context, req *tg.MessagesSendMediaRequest) error {
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

// SendMultiMedia sends message to peer.
func (s Sender) SendMultiMedia(ctx context.Context, req *tg.MessagesSendMultiMediaRequest) error {
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
