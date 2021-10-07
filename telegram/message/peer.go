package message

import (
	"context"

	"go.uber.org/atomic"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/telegram/message/peer"
	"github.com/nnqq/td/tg"
)

type resolvedCache atomic.Value

func (r *resolvedCache) Store(result tg.InputPeerClass) {
	r.Value.Store(result)
}

func (r *resolvedCache) Load() (result tg.InputPeerClass, ok bool) {
	result, ok = r.Value.Load().(tg.InputPeerClass)
	return
}

func (s *Sender) builder(promise peer.Promise, decorators []peer.PromiseDecorator) *RequestBuilder {
	once := &resolvedCache{}

	for _, decorator := range decorators {
		promise = decorator(promise)
	}
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
func (s *Sender) PeerPromise(p peer.Promise, decorators ...peer.PromiseDecorator) *RequestBuilder {
	return s.builder(p, decorators)
}

// To uses given peer to create new message builder.
func (s *Sender) To(p tg.InputPeerClass) *RequestBuilder {
	return s.PeerPromise(func(ctx context.Context) (tg.InputPeerClass, error) {
		return p, nil
	})
}

// Self creates a new message builder to send it to yourself.
// It means that message will be sent to your Saved Messages folder.
func (s *Sender) Self() *RequestBuilder {
	return s.To(&tg.InputPeerSelf{})
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
//	+13115552368
//	+1 (311) 555-0123
//	+1 311 555-6162
//
func (s *Sender) Resolve(from string, decorators ...peer.PromiseDecorator) *RequestBuilder {
	return s.builder(peer.Resolve(s.resolver, from), decorators)
}

// ResolvePhone uses given phone to create new peer promise.
// It resolves peer of message using given Resolver.
// Input example:
//
// 	+13115552368
// 	+1 (311) 555-0123
// 	+1 311 555-6162
//
// NB: ResolvePhone just deletes any non-digit symbols from phone argument.
// For now, Telegram sends contact number as string like "13115552368".
func (s *Sender) ResolvePhone(phone string, decorators ...peer.PromiseDecorator) *RequestBuilder {
	return s.builder(peer.ResolvePhone(s.resolver, phone), decorators)
}

// ResolveDomain uses given domain to create new message builder.
// It resolves peer of message using Sender's PeerResolver.
// Can has prefix with @ or not.
// Input examples:
//
//	@telegram
//	telegram
//
func (s *Sender) ResolveDomain(domain string, decorators ...peer.PromiseDecorator) *RequestBuilder {
	return s.builder(peer.ResolveDomain(s.resolver, domain), decorators)
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
func (s *Sender) ResolveDeeplink(link string, decorators ...peer.PromiseDecorator) *RequestBuilder {
	return s.builder(peer.ResolveDeeplink(s.resolver, link), decorators)
}

// PeerUpdate represents update which can be used to answer.
type PeerUpdate interface {
	GetPeer() tg.PeerClass
}

// Peer uses given peer update to create message for same chat.
func (s *Sender) Peer(uctx tg.Entities, upd PeerUpdate, decorators ...peer.PromiseDecorator) *RequestBuilder {
	entities := peer.EntitiesFromUpdate(uctx)
	return s.builder(func(ctx context.Context) (tg.InputPeerClass, error) {
		return entities.ExtractPeer(upd.GetPeer())
	}, decorators)
}

// AnswerableMessageUpdate represents update which can be used to answer.
type AnswerableMessageUpdate interface {
	GetMessage() tg.MessageClass
}

// Answer uses given message update to create message for same chat.
func (s *Sender) Answer(
	uctx tg.Entities,
	upd AnswerableMessageUpdate,
	decorators ...peer.PromiseDecorator,
) *RequestBuilder {
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
	}, decorators)
}

// Reply uses given message update to create message for same chat and create a reply.
// Shorthand for
//
// 	sender.Answer(uctx, upd).ReplyMsg(upd.GetMessage())
//
func (s *Sender) Reply(uctx tg.Entities, upd AnswerableMessageUpdate, decorators ...peer.PromiseDecorator) *Builder {
	return s.Answer(uctx, upd, decorators...).ReplyMsg(upd.GetMessage())
}
