// Package widget contains helpers for Telegram HTML widget messages.
//
// A widget is an interactive piece of HTML that a bot sends as message media
// (inputMediaWidget) and Telegram clients render inside the chat bubble, sized
// to its content. If sent with allow_queries, the widget can send queries to
// the bot (updateBotWidgetQuery), which the bot answers with
// messages.setBotWidgetQueryAnswer. The bot can also replace widget data of an
// already sent message (messages.editMessageWidgetData) or, for shared
// widgets, of every message showing it at once (messages.editWidgetData).
//
// Outgoing HTML is sent as is. [ValidateHTML] checks it client-side: it must
// be 1-[MaxHTMLLength] bytes of UTF-8. Telegram also requires the HTML to
// compress (zlib, best compression) to at most [MaxCompressedHTMLLength]
// bytes; that limit is enforced by the server, and [ValidateCompressed] only
// approximates it for callers who want an early check.
//
// Incoming messageMediaWidget carries the HTML zlib-compressed in its payload
// field; use [DecodePayload] or [HTML] to get the source back.
//
// Use telegram/message (message.Widget, message.SharedWidget) to send widgets
// and telegram/message/inline (inline.MessageWidget) for inline results.
package widget

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"
	"unicode/utf8"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

const (
	// MaxHTMLLength is the maximum length of widget HTML in bytes.
	MaxHTMLLength = 32768
	// MaxCompressedHTMLLength is the maximum length of widget HTML in bytes
	// after zlib compression with the maximum compression level.
	MaxCompressedHTMLLength = 4096

	// maxDecodedLength limits decompressed payload size as a safety net
	// against decompression bombs. The server enforces the real limits.
	maxDecodedLength = 4 << 20
)

var (
	// ErrEmptyHTML is returned when widget HTML is empty.
	ErrEmptyHTML = errors.New("widget HTML must be non-empty")
	// ErrHTMLTooLong is returned when widget HTML exceeds MaxHTMLLength bytes.
	ErrHTMLTooLong = errors.New("widget HTML is too long")
	// ErrHTMLCompressedTooLong is returned when widget HTML compressed with
	// zlib exceeds MaxCompressedHTMLLength bytes.
	ErrHTMLCompressedTooLong = errors.New("compressed widget HTML is too long")
	// ErrInvalidUTF8 is returned when widget HTML or decoded payload is not
	// valid UTF-8.
	ErrInvalidUTF8 = errors.New("widget HTML must be encoded in UTF-8")
	// ErrInvalidSize is returned when widget width or height is negative.
	ErrInvalidSize = errors.New("invalid widget size")
	// ErrInvalidWidgetID is returned when shared widget identifier is zero.
	ErrInvalidWidgetID = errors.New("invalid widget ID")
)

// CompressedSize returns length of html compressed by zlib with the maximum
// compression level, as used by the MaxCompressedHTMLLength limit.
//
// Note that Go compress/flate output is not byte-identical to C zlib, so the
// result may differ from the server-side value by a few bytes.
func CompressedSize(html string) (int, error) {
	payload, err := EncodePayload(html)
	if err != nil {
		return 0, err
	}
	return len(payload), nil
}

// EncodePayload compresses html with zlib at the maximum compression level.
//
// The result has the same format as messageMediaWidget.payload, which is
// useful for tests and server implementations. Clients send raw HTML in
// inputMediaWidget.html, so EncodePayload is not needed to send widgets.
func EncodePayload(html string) ([]byte, error) {
	var buf bytes.Buffer
	w, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		return nil, errors.Wrap(err, "create zlib writer")
	}
	if _, err := io.WriteString(w, html); err != nil {
		return nil, errors.Wrap(err, "compress")
	}
	if err := w.Close(); err != nil {
		return nil, errors.Wrap(err, "close zlib writer")
	}
	return buf.Bytes(), nil
}

// ValidateHTML checks widget HTML against the client-side limits:
// 1-MaxHTMLLength bytes of valid UTF-8 (td_api.tl inputMessageWidget).
//
// It does not check the compressed size limit, see ValidateCompressed.
// Returned error matches (errors.Is) one of ErrEmptyHTML, ErrHTMLTooLong or
// ErrInvalidUTF8.
func ValidateHTML(html string) error {
	switch {
	case html == "":
		return ErrEmptyHTML
	case len(html) > MaxHTMLLength:
		return errors.Wrapf(ErrHTMLTooLong, "%d bytes, limit is %d", len(html), MaxHTMLLength)
	case !utf8.ValidString(html):
		return ErrInvalidUTF8
	}
	return nil
}

// ValidateCompressed reports whether html compressed by zlib with the maximum
// compression level exceeds MaxCompressedHTMLLength bytes, returning error
// matching ErrHTMLCompressedTooLong if so.
//
// This is an approximation: Go compress/flate output is not byte-identical to
// C zlib, so near the limit the result may differ from the server, which is
// authoritative. Neither tdlib nor the message builders of this module enforce
// this limit client-side.
func ValidateCompressed(html string) error {
	size, err := CompressedSize(html)
	if err != nil {
		return errors.Wrap(err, "compute compressed size")
	}
	if size > MaxCompressedHTMLLength {
		return errors.Wrapf(ErrHTMLCompressedTooLong,
			"%d bytes, limit is %d", size, MaxCompressedHTMLLength)
	}
	return nil
}

// ValidateSize checks widget width and height. Zero means full bubble width
// (or unfixed height); negative values are rejected, as tdlib does.
func ValidateSize(width, height int) error {
	if width < 0 || height < 0 {
		return errors.Wrapf(ErrInvalidSize, "%dx%d", width, height)
	}
	return nil
}

// DecodePayload decompresses messageMediaWidget (or botInlineMessageWidget)
// payload into widget HTML.
//
// The server sends zlib-compressed HTML; gzip is accepted for parity with
// tdlib (gzdecode) and Telegram Desktop (Tgw::UnpackPayload), which both use
// inflateInit2(MAX_WBITS + 32). As there, only the first gzip member is read
// and trailing bytes are ignored. The result must be non-empty valid UTF-8,
// otherwise the payload is invalid, as in tdlib.
func DecodePayload(payload []byte) (string, error) {
	if len(payload) == 0 {
		return "", errors.New("empty payload")
	}
	var (
		r   io.ReadCloser
		err error
	)
	if len(payload) >= 2 && payload[0] == 0x1f && payload[1] == 0x8b {
		var gr *gzip.Reader
		if gr, err = gzip.NewReader(bytes.NewReader(payload)); err == nil {
			gr.Multistream(false)
			r = gr
		}
	} else {
		r, err = zlib.NewReader(bytes.NewReader(payload))
	}
	if err != nil {
		return "", errors.Wrap(err, "open payload")
	}
	defer func() { _ = r.Close() }()

	data, err := io.ReadAll(io.LimitReader(r, maxDecodedLength+1))
	if err != nil {
		return "", errors.Wrap(err, "decompress payload")
	}
	if len(data) > maxDecodedLength {
		return "", errors.Errorf("decompressed payload exceeds %d bytes", maxDecodedLength)
	}
	if len(data) == 0 {
		return "", ErrEmptyHTML
	}
	if !utf8.Valid(data) {
		return "", ErrInvalidUTF8
	}
	return string(data), nil
}

// HTML returns HTML source of the widget from message media.
func HTML(m *tg.MessageMediaWidget) (string, error) {
	if m == nil {
		return "", errors.New("nil widget media")
	}
	return DecodePayload(m.Payload)
}
