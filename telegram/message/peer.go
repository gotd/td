package message

import (
	"context"

	"go.uber.org/atomic"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

type resolvedCache atomic.Value

func (r *resolvedCache) Store(result tg.InputPeerClass) {
	r.Value.Store(result)
}

func (r *resolvedCache) Load() (result tg.InputPeerClass, ok bool) {
	result, ok = r.Value.Load().(tg.InputPeerClass)
	return
}

func (s *Sender) builder(promise peer.Promise) *RequestBuilder {
	once := &resolvedCache{}

	return &RequestBuilder{
		Builder: Builder{
			sender: s,
			peer: func(ctx context.Context) (r tg.InputPeerClass, err error) {
				if v, ok := once.Load(); ok {
					return v, nil
				}
				defer func() {
					if err == nil && r != nil {
						once.Store(r)
					}
				}()

				return promise(ctx)
			},
		},
	}
}

// PeerPromise uses given peer promise to create new message builder.
func (s *Sender) PeerPromise(p peer.Promise) *RequestBuilder {
	return s.builder(p)
}

// Peer uses given peer to create new message builder.
func (s *Sender) Peer(p tg.InputPeerClass) *RequestBuilder {
	return s.PeerPromise(func(ctx context.Context) (tg.InputPeerClass, error) {
		return p, nil
	})
}

// Self creates a new message builder to send it to yourself.
// It means that message will be sent to your Saved Messages folder.
func (s *Sender) Self() *RequestBuilder {
	return s.Peer(&tg.InputPeerSelf{})
}

// AsInputPeer returns resolve result as InputPeerClass.
func (b *RequestBuilder) AsInputPeer(ctx context.Context) (tg.InputPeerClass, error) {
	return b.peer(ctx)
}

// Resolve uses given text to create new message builder.
// It resolves peer of message using Sender's PeerResolver.
// Input examples:
//
//	@telegram
//	telegram
//	t.me/telegram
//	https://t.me/telegram
//	tg:resolve?domain=telegram
//	tg://resolve?domain=telegram
//
func (s *Sender) Resolve(from string) *RequestBuilder {
	return s.builder(peer.Resolve(s.resolver, from))
}

// ResolveDomain uses given domain to create new message builder.
// It resolves peer of message using Sender's PeerResolver.
// Can has prefix with @ or not.
// Input examples:
//
//	@telegram
//	telegram
//
func (s *Sender) ResolveDomain(domain string) *RequestBuilder {
	return s.builder(peer.ResolveDomain(s.resolver, domain))
}

// ResolveDeeplink uses given deeplink to create new message builder.
// Deeplink is a URL like https://t.me/telegram.
// It resolves peer of message using Sender's PeerResolver.
// Input examples:
//
//	t.me/telegram
//	https://t.me/telegram
//	tg:resolve?domain=telegram
//	tg://resolve?domain=telegram
//
func (s *Sender) ResolveDeeplink(deeplink string) *RequestBuilder {
	return s.builder(peer.ResolveDeeplink(s.resolver, deeplink))
}

// AnswerableMessageUpdate represents update which can be used to answer.
type AnswerableMessageUpdate interface {
	GetMessage() tg.MessageClass
	GetPts() int
}

// Answer uses given message update to create message for same chat.
func (s *Sender) Answer(uctx tg.UpdateContext, upd AnswerableMessageUpdate) *RequestBuilder {
	entities := peer.EntitiesFromUpdate(uctx)
	return s.builder(func(ctx context.Context) (tg.InputPeerClass, error) {
		updMsg := upd.GetMessage()
		msg, ok := updMsg.AsNotEmpty()
		if !ok {
			emptyMsg, ok := updMsg.(*tg.MessageEmpty)
			if !ok {
				return nil, xerrors.Errorf("unexpected type %T", updMsg)
			}

			p, ok := emptyMsg.GetPeerID()
			if !ok {
				return nil, xerrors.Errorf("got %T with empty PeerID", updMsg)
			}

			return entities.ExtractPeer(p)
		}

		return entities.ExtractPeer(msg.GetPeerID())
	})
}

// Reply uses given message update to create message for same chat and create a reply.
// Shorthand for
//
// 	sender.Answer(uctx, upd).ReplyMsg(upd.GetMessage())
//
func (s *Sender) Reply(uctx tg.UpdateContext, upd AnswerableMessageUpdate) *Builder {
	return s.Answer(uctx, upd).ReplyMsg(upd.GetMessage())
}
