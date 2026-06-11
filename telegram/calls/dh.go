package calls

import (
	"context"
	"crypto/sha1" //nolint:gosec // SHA-1 is mandated by the Telegram call-key fingerprint scheme.
	"crypto/sha256"
	"io"
	"math/big"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// keySize is the byte length of the DH prime, exponents and shared key.
//
// Telegram phone calls use a 2048-bit (256-byte) group.
const keySize = 256

var (
	bigOne = big.NewInt(1)
	bigTwo = big.NewInt(2)
)

// dhConfig holds the Diffie-Hellman group used for a call, as returned by
// messages.getDhConfig.
type dhConfig struct {
	g int
	p *big.Int
}

// getDHConfig fetches and validates the current DH configuration.
func getDHConfig(ctx context.Context, api *tg.Client) (*dhConfig, []byte, error) {
	cfg, err := api.MessagesGetDhConfig(ctx, &tg.MessagesGetDhConfigRequest{
		Version:      0,
		RandomLength: keySize,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "get dh config")
	}
	obj, ok := cfg.(*tg.MessagesDhConfig)
	if !ok {
		// messages.dhConfigNotModified is only returned when a non-zero
		// version is supplied; with version 0 we always get the full config.
		return nil, nil, errors.Errorf("unexpected dh config %T", cfg)
	}
	dh := &dhConfig{g: obj.G, p: new(big.Int).SetBytes(obj.P)}
	if err := dh.validate(); err != nil {
		return nil, nil, err
	}
	return dh, obj.Random, nil
}

// validate performs the sanity checks required before using the group.
func (dh *dhConfig) validate() error {
	if dh.p.BitLen() != keySize*8 {
		return errors.Errorf("dh prime is not %d-bit (got %d)", keySize*8, dh.p.BitLen())
	}
	if dh.p.Sign() <= 0 {
		return errors.New("dh prime is not positive")
	}
	if dh.g < 2 || dh.g > 7 {
		return errors.Errorf("dh generator out of range: %d", dh.g)
	}
	return nil
}

// randomExp returns a random secret exponent in [2, p-2] together with
// g^exp mod p, mixing the caller-provided server random with locally read
// randomness as recommended by Telegram.
func (dh *dhConfig) randomExp(rnd io.Reader, serverRandom []byte) (exp, pub *big.Int, err error) {
	buf := make([]byte, keySize)
	for range 16 {
		if _, err := io.ReadFull(rnd, buf); err != nil {
			return nil, nil, errors.Wrap(err, "read random")
		}
		// Mix in the server-provided random so neither side fully controls it.
		if len(serverRandom) == len(buf) {
			for i := range buf {
				buf[i] ^= serverRandom[i]
			}
		}
		exp = new(big.Int).SetBytes(buf)
		// exp = 2 + (exp mod (p-3)) lands in [2, p-2].
		exp.Mod(exp, new(big.Int).Sub(dh.p, big.NewInt(3)))
		exp.Add(exp, bigTwo)

		pub = new(big.Int).Exp(big.NewInt(int64(dh.g)), exp, dh.p)
		if dh.checkValue(pub) == nil {
			return exp, pub, nil
		}
	}
	return nil, nil, errors.New("failed to generate valid dh exponent")
}

// checkValue ensures a received or computed public value lies in (1, p-1),
// rejecting the small-subgroup confinement values.
func (dh *dhConfig) checkValue(v *big.Int) error {
	if v.Cmp(bigOne) <= 0 {
		return errors.New("dh value too small")
	}
	if v.Cmp(new(big.Int).Sub(dh.p, bigOne)) >= 0 {
		return errors.New("dh value too large")
	}
	return nil
}

// computeKey derives the 256-byte shared key and its 64-bit fingerprint from
// the peer's public value and our secret exponent.
func (dh *dhConfig) computeKey(peerPub []byte, exp *big.Int) (key []byte, fingerprint int64, err error) {
	pub := new(big.Int).SetBytes(peerPub)
	if err := dh.checkValue(pub); err != nil {
		return nil, 0, errors.Wrap(err, "peer public value")
	}
	key = pad(new(big.Int).Exp(pub, exp, dh.p))
	return key, keyFingerprint(key), nil
}

// keyFingerprint is the low 64 bits of SHA1(key), little-endian, matching the
// key_fingerprint sent in phone.confirmCall.
func keyFingerprint(key []byte) int64 {
	sum := sha1.Sum(key) //nolint:gosec // Mandated by the protocol.
	tail := sum[len(sum)-8:]
	var fp int64
	for i := range 8 {
		fp |= int64(tail[i]) << (uint(i) * 8)
	}
	return fp
}

// gAHash returns SHA256(g_a), the commitment the caller publishes before the
// callee reveals g_b.
func gAHash(gA []byte) []byte {
	sum := sha256.Sum256(gA)
	return sum[:]
}

// pad left-pads the big-endian bytes of n to keySize bytes.
func pad(n *big.Int) []byte {
	b := n.Bytes()
	if len(b) >= keySize {
		return b
	}
	out := make([]byte, keySize)
	copy(out[keySize-len(b):], b)
	return out
}
