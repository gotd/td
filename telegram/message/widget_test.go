package message

import (
	"context"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/message/widget"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

// wireRoundTrip encodes v and decodes it into out, so tests check what is
// actually sent, including flags.
func wireRoundTrip(t *testing.T, v bin.Encoder, out bin.Decoder) {
	t.Helper()
	var buf bin.Buffer
	require.NoError(t, v.Encode(&buf))
	require.NoError(t, out.Decode(&buf))
}

// poorlyCompressibleHTML returns deterministic ASCII text that compresses to
// roughly 3/4 of its length.
func poorlyCompressibleHTML(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789<>"
	r := rand.New(rand.NewPCG(1, 2)) // #nosec G404
	var b strings.Builder
	for range n {
		b.WriteByte(alphabet[r.IntN(len(alphabet))])
	}
	return b.String()
}

func expectWidgetMedia(t *testing.T, mock *tgmock.Mock, check func(req *tg.MessagesSendMediaRequest)) {
	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesSendMediaRequest)
		require.True(t, ok)
		var decoded tg.MessagesSendMediaRequest
		wireRoundTrip(t, req, &decoded)
		check(&decoded)
	}).ThenResult(&tg.Updates{})
}

func TestWidget(t *testing.T) {
	ctx := context.Background()
	const html = `<div id="app">0</div>`

	t.Run("Full", func(t *testing.T) {
		sender, mock := testSender(t)
		expectWidgetMedia(t, mock, func(req *tg.MessagesSendMediaRequest) {
			require.True(t, req.InvertMedia)
			require.Equal(t, "Counter", req.Message)
			require.Len(t, req.Entities, 1)

			m, ok := req.Media.(*tg.InputMediaWidget)
			require.True(t, ok)
			require.True(t, m.AllowQueries)
			require.True(t, m.SharedWidget)
			require.Equal(t, 320, m.Width)
			require.Equal(t, 0, m.Height)
			require.Equal(t, html, m.HTML)
			require.Equal(t, `{"n":0}`, m.PublicData)
			require.Equal(t, "chat", m.ChatData)
		})
		_, err := sender.Self().InvertMedia().Media(ctx,
			Widget(html, styling.Bold("Counter")).
				Size(320, 0).
				PublicData(`{"n":0}`).
				ChatData("chat").
				AllowQueries(true).
				Shared(true),
		)
		require.NoError(t, err)
	})
	t.Run("Plain", func(t *testing.T) {
		sender, mock := testSender(t)
		expectSendMedia(t, &tg.InputMediaWidget{HTML: html}, mock)
		_, err := sender.Self().Widget(ctx, html)
		require.NoError(t, err)
	})
	t.Run("Validation", func(t *testing.T) {
		sender, _ := testSender(t)
		for _, tt := range []struct {
			name string
			opt  MediaOption
			err  error
		}{
			{"Empty", Widget(""), widget.ErrEmptyHTML},
			{"TooLong", Widget(strings.Repeat("a", widget.MaxHTMLLength+1)), widget.ErrHTMLTooLong},
			{"NegativeSize", Widget(html).Size(-1, 10), widget.ErrInvalidSize},
			{"InvalidUTF8", Widget("\xff"), widget.ErrInvalidUTF8},
			{"SharedZeroID", SharedWidget(0, ""), widget.ErrInvalidWidgetID},
		} {
			t.Run(tt.name, func(t *testing.T) {
				// No RPC is expected: mock fails on unexpected calls.
				_, err := sender.Self().Media(ctx, tt.opt)
				require.ErrorIs(t, err, tt.err)
			})
		}
	})
	t.Run("OverCompressedLimit", func(t *testing.T) {
		// The compressed size limit is enforced by the server only.
		sender, mock := testSender(t)
		html := poorlyCompressibleHTML(6000)
		size, err := widget.CompressedSize(html)
		require.NoError(t, err)
		require.Greater(t, size, widget.MaxCompressedHTMLLength)
		expectSendMedia(t, &tg.InputMediaWidget{HTML: html}, mock)
		_, err = sender.Self().Widget(ctx, html)
		require.NoError(t, err)
	})
	t.Run("MaxLength", func(t *testing.T) {
		sender, mock := testSender(t)
		long := strings.Repeat("a", widget.MaxHTMLLength)
		expectSendMedia(t, &tg.InputMediaWidget{HTML: long}, mock)
		_, err := sender.Self().Widget(ctx, long)
		require.NoError(t, err)
	})
}

func TestSharedWidget(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	expectSendMediaAndText(t, &tg.InputMediaWidgetID{WidgetID: 42, ChatData: "c"}, mock, "caption")
	_, err := sender.Self().SharedWidget(ctx, 42, "c", styling.Plain("caption"))
	require.NoError(t, err)

	expectWidgetMedia(t, mock, func(req *tg.MessagesSendMediaRequest) {
		m, ok := req.Media.(*tg.InputMediaWidgetID)
		require.True(t, ok)
		require.Equal(t, int64(43), m.WidgetID)
		_, ok = m.GetChatData()
		require.False(t, ok)
	})
	_, err = sender.Self().Media(ctx, SharedWidget(43, ""))
	require.NoError(t, err)
}

func TestEditMessageBuilder_WidgetData(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	expect := func(check func(req *tg.MessagesEditMessageWidgetDataRequest)) {
		mock.ExpectFunc(func(b bin.Encoder) {
			req, ok := b.(*tg.MessagesEditMessageWidgetDataRequest)
			require.True(t, ok)
			var decoded tg.MessagesEditMessageWidgetDataRequest
			wireRoundTrip(t, req, &decoded)
			require.Equal(t, &tg.InputPeerSelf{}, decoded.Peer)
			require.Equal(t, 10, decoded.ID)
			check(&decoded)
		}).ThenResult(&tg.Updates{})
	}

	expect(func(req *tg.MessagesEditMessageWidgetDataRequest) {
		v, ok := req.GetPublicData()
		require.True(t, ok)
		require.Equal(t, "pub", v)
		_, ok = req.GetChatData()
		require.False(t, ok)
	})
	_, err := sender.Self().Edit(10).WidgetData(ctx, WidgetPublicData("pub"))
	require.NoError(t, err)

	// Empty string must still be sent to clear the data.
	expect(func(req *tg.MessagesEditMessageWidgetDataRequest) {
		_, ok := req.GetPublicData()
		require.False(t, ok)
		v, ok := req.GetChatData()
		require.True(t, ok)
		require.Empty(t, v)
	})
	_, err = sender.Self().Edit(10).WidgetData(ctx, WidgetChatData(""))
	require.NoError(t, err)

	_, err = sender.Self().Edit(10).WidgetData(ctx)
	require.ErrorIs(t, err, ErrNoWidgetData)

	mock.ExpectCall(&tg.MessagesEditMessageWidgetDataRequest{
		Flags: 3, Peer: &tg.InputPeerSelf{}, ID: 10, PublicData: "x", ChatData: "y",
	}).ThenRPCErr(testRPCError())
	_, err = sender.Self().Edit(10).WidgetData(ctx, WidgetPublicData("x"), WidgetChatData("y"))
	require.Error(t, err)
}

func TestSender_EditInlineWidgetData(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)
	id := &tg.InputBotInlineMessageID64{DCID: 2, OwnerID: 1, ID: 3, AccessHash: 4}

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesEditInlineBotMessageWidgetDataRequest)
		require.True(t, ok)
		var decoded tg.MessagesEditInlineBotMessageWidgetDataRequest
		wireRoundTrip(t, req, &decoded)
		require.Equal(t, id, decoded.ID)
		v, ok := decoded.GetPublicData()
		require.True(t, ok)
		require.Empty(t, v)
		v, ok = decoded.GetChatData()
		require.True(t, ok)
		require.Equal(t, "chat", v)
	}).ThenTrue()
	require.NoError(t, sender.EditInlineWidgetData(ctx, id, WidgetPublicData(""), WidgetChatData("chat")))

	require.ErrorIs(t, sender.EditInlineWidgetData(ctx, id), ErrNoWidgetData)
}

func TestSender_EditSharedWidgetData(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectFunc(func(b bin.Encoder) {
		req, ok := b.(*tg.MessagesEditWidgetDataRequest)
		require.True(t, ok)
		var decoded tg.MessagesEditWidgetDataRequest
		wireRoundTrip(t, req, &decoded)
		require.Equal(t, int64(7), decoded.WidgetID)
		require.Equal(t, "data", decoded.PublicData)
		_, ok = decoded.GetCacheTime()
		require.False(t, ok)
	}).ThenTrue()
	require.NoError(t, sender.EditSharedWidgetData(ctx, 7, "data", 0))

	for _, d := range []time.Duration{5 * time.Second, time.Hour} {
		mock.ExpectCall(&tg.MessagesEditWidgetDataRequest{
			Flags: 1, WidgetID: 7, PublicData: "data", CacheTime: int(d.Seconds()),
		}).ThenTrue()
		require.NoError(t, sender.EditSharedWidgetData(ctx, 7, "data", d))
	}

	for _, d := range []time.Duration{time.Second, 4 * time.Second, time.Hour + time.Second, -time.Second} {
		require.ErrorIs(t, sender.EditSharedWidgetData(ctx, 7, "data", d), ErrInvalidWidgetCacheTime)
	}
	require.ErrorIs(t, sender.EditSharedWidgetData(ctx, 0, "data", 0), widget.ErrInvalidWidgetID)
}

func TestSender_AnswerWidgetQuery(t *testing.T) {
	ctx := context.Background()
	sender, mock := testSender(t)

	mock.ExpectCall(&tg.MessagesSetBotWidgetQueryAnswerRequest{
		QueryID: 10, Result: `{"ok":true}`,
	}).ThenTrue()
	require.NoError(t, sender.AnswerWidgetQuery(ctx, 10, `{"ok":true}`))

	mock.ExpectCall(&tg.MessagesSetBotWidgetQueryAnswerRequest{QueryID: 11}).ThenRPCErr(testRPCError())
	require.Error(t, sender.AnswerWidgetQuery(ctx, 11, ""))
}
