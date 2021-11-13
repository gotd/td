package message

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

// AsInputPeer returns resolve result as InputPeerClass.
func (b *RequestBuilder) AsInputPeer(ctx context.Context) (tg.InputPeerClass, error) {
	return b.peer(ctx)
}

// AsInputUserClass returns resolve result as InputUserClass.
func (b *RequestBuilder) AsInputUserClass(ctx context.Context) (tg.InputUserClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, err
	}

	user, ok := peer.ToInputUser(p)
	if !ok {
		return nil, errors.Errorf("unexpected type %T", p)
	}
	return user, nil
}

// AsInputUser returns resolve result as InputUser.
func (b *RequestBuilder) AsInputUser(ctx context.Context) (*tg.InputUser, error) {
	user, err := b.AsInputUserClass(ctx)
	if err != nil {
		return nil, err
	}

	userID, ok := user.(*tg.InputUser)
	if !ok {
		return nil, errors.Errorf("unexpected type %T", user)
	}
	return userID, nil
}

// AsInputChannelClass returns resolve result as tg.NotEmptyInputChannel.
func (b *RequestBuilder) AsInputChannelClass(ctx context.Context) (tg.InputChannelClass, error) {
	p, err := b.peer(ctx)
	if err != nil {
		return nil, err
	}

	channel, ok := peer.ToInputChannel(p)
	if !ok {
		return nil, errors.Errorf("unexpected type %T", p)
	}
	return channel, nil
}

// AsInputChannel returns resolve result as InputChannel.
func (b *RequestBuilder) AsInputChannel(ctx context.Context) (*tg.InputChannel, error) {
	channel, err := b.AsInputChannelClass(ctx)
	if err != nil {
		return nil, err
	}

	channelID, ok := channel.(*tg.InputChannel)
	if !ok {
		return nil, errors.Errorf("unexpected type %T", channel)
	}
	return channelID, nil
}
