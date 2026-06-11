package calls_test

import (
	"context"

	"github.com/pion/webrtc/v4"

	"github.com/gotd/td/telegram/calls"
	"github.com/gotd/td/tg"
)

// This example shows how to wire the call client to an update dispatcher and
// place an outgoing call. Error handling is elided for brevity.
func ExampleClient() {
	// In a real program the dispatcher is passed to telegram.Options as the
	// UpdateHandler, and api wraps the connected client.
	dispatcher := tg.NewUpdateDispatcher()
	var api *tg.Client // = tg.NewClient(client)

	c := calls.NewClient(api, calls.Options{})
	c.Register(dispatcher)

	// Answer incoming calls.
	c.OnIncoming(func(in *calls.IncomingCall) {
		conn, err := in.Accept(context.Background())
		if err != nil {
			return
		}
		conn.OnTrack(func(remote *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
			_ = remote // decode incoming media
		})
	})

	// Place an outgoing call to a resolved input user.
	var user tg.InputUserClass
	conn, err := c.Request(context.Background(), user)
	if err != nil {
		return
	}
	conn.OnConnected(func() {
		// Write Opus RTP packets to conn.AudioTrack().
	})
}
