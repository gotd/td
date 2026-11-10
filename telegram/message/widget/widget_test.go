package widget

import (
	"encoding/base64"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

const fixtureHTML = `<!doctype html><html><body><h1>Hello, мир!</h1>` +
	`<button onclick="tg.query()">Go</button></body></html>`

// Fixtures are produced by C zlib (the library tdlib and the server use):
//
//	zlib.compress(html, 9), zlib.compress(html, 6), gzip.compress(html, 9, mtime=0)
var fixtures = map[string]string{
	"zlib9": "eNqzUUzJTy6pLEhVyCjJzbGzgZBJ+SmVQLahnUdqTk6+jsKFPRd2XGxQtNEHCtkklZaU5Ocp5Ocl52QmZ9sql" +
		"aTrFZamFlVqaCrZuefb6EPk7YAMsCn6YCMB3YEltQ==",
	"zlib6": "eJyzUUzJTy6pLEhVyCjJzbGzgZBJ+SmVQLahnUdqTk6+jsKFPRd2XGxQtNEHCtkklZaU5Ocp5Ocl52QmZ9sql" +
		"aTrFZamFlVqaCrZuefb6EPk7YAMsCn6YCMB3YEltQ==",
	"gzip": "H4sIAAAAAAAC/7NRTMlPLqksSFXIKMnNsbOBkEn5KZVAtqGdR2pOTr6OwoU9F3ZcbFC00QcK2SSVlpTk5ynk5yXnZ" +
		"CZn2yqVpOsVlqYWVWpoKtm559voQ+TtgAywKfpgIwERK0RjaAAAAA==",
}

func TestDecodePayloadFixtures(t *testing.T) {
	for name, b64 := range fixtures {
		t.Run(name, func(t *testing.T) {
			payload, err := base64.StdEncoding.DecodeString(b64)
			require.NoError(t, err)

			html, err := DecodePayload(payload)
			require.NoError(t, err)
			require.Equal(t, fixtureHTML, html)

			html, err = HTML(&tg.MessageMediaWidget{Payload: payload})
			require.NoError(t, err)
			require.Equal(t, fixtureHTML, html)
		})
	}
}

func TestPayloadRoundTrip(t *testing.T) {
	for _, html := range []string{
		"a",
		fixtureHTML,
		strings.Repeat("<div>widget</div>", 1000),
	} {
		payload, err := EncodePayload(html)
		require.NoError(t, err)
		// zlib header, best compression.
		require.Equal(t, []byte{0x78, 0xda}, payload[:2])

		decoded, err := DecodePayload(payload)
		require.NoError(t, err)
		require.Equal(t, html, decoded)
	}
}

func TestDecodePayloadInvalid(t *testing.T) {
	_, err := DecodePayload(nil)
	require.Error(t, err)

	_, err = DecodePayload([]byte("<html></html>"))
	require.Error(t, err, "raw html is not a valid payload")

	payload, err := EncodePayload("")
	require.NoError(t, err)
	_, err = DecodePayload(payload)
	require.ErrorIs(t, err, ErrEmptyHTML)

	payload, err = EncodePayload("\xff\xfe")
	require.NoError(t, err)
	_, err = DecodePayload(payload)
	require.ErrorIs(t, err, ErrInvalidUTF8)

	valid, err := EncodePayload(fixtureHTML)
	require.NoError(t, err)
	_, err = DecodePayload(valid[:len(valid)-6])
	require.Error(t, err, "truncated payload")

	bomb, err := EncodePayload(strings.Repeat("a", maxDecodedLength+1))
	require.NoError(t, err)
	_, err = DecodePayload(bomb)
	require.Error(t, err)

	_, err = HTML(nil)
	require.Error(t, err)
}

// randomHTML returns deterministic poorly compressible ASCII text.
func randomHTML(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789<>"
	r := rand.New(rand.NewPCG(1, 2)) // #nosec G404
	var b strings.Builder
	for range n {
		b.WriteByte(alphabet[r.IntN(len(alphabet))])
	}
	return b.String()
}

func TestValidateHTML(t *testing.T) {
	require.ErrorIs(t, ValidateHTML(""), ErrEmptyHTML)
	require.ErrorIs(t, ValidateHTML("<b>\xff</b>"), ErrInvalidUTF8)
	require.NoError(t, ValidateHTML("x"))
	require.NoError(t, ValidateHTML(fixtureHTML))

	t.Run("Length", func(t *testing.T) {
		// Highly compressible, so only the raw length limit matters.
		require.NoError(t, ValidateHTML(strings.Repeat("a", MaxHTMLLength)))
		err := ValidateHTML(strings.Repeat("a", MaxHTMLLength+1))
		require.ErrorIs(t, err, ErrHTMLTooLong)
		require.False(t, errors.Is(err, ErrHTMLCompressedTooLong))
	})
	t.Run("CompressedNotChecked", func(t *testing.T) {
		html := randomHTML(MaxHTMLLength)
		require.ErrorIs(t, ValidateCompressed(html), ErrHTMLCompressedTooLong)
		require.NoError(t, ValidateHTML(html), "compressed size is enforced by the server only")
	})
}

func TestValidateCompressed(t *testing.T) {
	require.NoError(t, ValidateCompressed(fixtureHTML))
	require.NoError(t, ValidateCompressed(strings.Repeat("a", MaxHTMLLength)))

	t.Run("Boundary", func(t *testing.T) {
		src := randomHTML(MaxHTMLLength)
		compressed := func(i int) int {
			size, err := CompressedSize(src[:i])
			require.NoError(t, err)
			return size
		}
		// Compressed size of a random prefix grows (almost) monotonically,
		// so binary search the boundary and then scan around it for a prefix
		// that compresses to exactly the limit while the next one exceeds it.
		lo, hi := 0, len(src)
		for lo < hi {
			mid := (lo + hi) / 2
			if compressed(mid) > MaxCompressedHTMLLength {
				hi = mid
			} else {
				lo = mid + 1
			}
		}
		n := -1
		for i := lo - 64; i < lo+64; i++ {
			if compressed(i) == MaxCompressedHTMLLength && compressed(i+1) > MaxCompressedHTMLLength {
				n = i
				break
			}
		}
		require.Positive(t, n, "no prefix compresses to exactly the limit")

		require.NoError(t, ValidateCompressed(src[:n]))
		size, err := CompressedSize(src[:n+1])
		require.NoError(t, err)
		require.Greater(t, size, MaxCompressedHTMLLength)
		require.ErrorIs(t, ValidateCompressed(src[:n+1]), ErrHTMLCompressedTooLong)
		require.Less(t, n, MaxHTMLLength)
	})
}

func TestValidateSize(t *testing.T) {
	require.NoError(t, ValidateSize(0, 0))
	require.NoError(t, ValidateSize(300, 200))
	require.ErrorIs(t, ValidateSize(-1, 0), ErrInvalidSize)
	require.ErrorIs(t, ValidateSize(0, -1), ErrInvalidSize)
}
