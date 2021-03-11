package message

import (
	"context"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

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

