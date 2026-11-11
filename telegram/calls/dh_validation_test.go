package calls

import (
	"context"
	"crypto/rand"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
)

func dhConfigResult(g int, p *big.Int) *tg.MessagesDhConfig {
	return &tg.MessagesDhConfig{G: g, P: p.Bytes(), Random: make([]byte, keySize)}
}

func TestGetDHConfigRejectsInvalidGroups(t *testing.T) {
	ctx := context.Background()

	// 2048-bit, but divisible by 3.
	composite := new(big.Int).Lsh(bigOne, keySize*8-1)
	composite.Add(composite, bigOne)

	// A 2048-bit prime that is (almost surely) not a safe prime; g = 4 has
	// no p mod 4g condition, so only the (p-1)/2 primality check rejects it.
	notSafe, err := rand.Prime(rand.Reader, keySize*8)
	require.NoError(t, err)

	rfc := testGroup(t).p

	for _, tt := range []struct {
		name string
		g    int
		p    *big.Int
	}{
		{"Composite", 3, composite},
		{"NotSafePrime", 4, notSafe},
		{"GeneratorOutOfRange", 8, rfc},
		{"ShortPrime", 2, new(big.Int).Rsh(rfc, 8)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mock := tgmock.NewRequire(t)
			mock.Expect().ThenResult(dhConfigResult(tt.g, tt.p))
			_, _, err := getDHConfig(ctx, tg.NewClient(mock))
			require.Error(t, err)
		})
	}

	// g = 2 needs p mod 8 = 7; the RFC 3526 prime satisfies it, g = 3 needs
	// p mod 3 = 2.
	g3 := new(big.Int).Mod(rfc, big.NewInt(3)).Int64() == 2
	mock := tgmock.NewRequire(t)
	mock.Expect().ThenResult(dhConfigResult(3, rfc))
	_, _, err = getDHConfig(ctx, tg.NewClient(mock))
	if g3 {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
	}
}

func TestDHValidateCached(t *testing.T) {
	dh := testGroup(t)
	require.NoError(t, dh.validate())
	start := time.Now()
	for range 100 {
		require.NoError(t, dh.validate())
	}
	// 100 primality tests of a 2048-bit safe prime take seconds.
	require.Less(t, time.Since(start), time.Second)
}

func TestDHCheckValueSafetyRange(t *testing.T) {
	dh := testGroup(t)
	two := big.NewInt(2)
	pMinusTwo := new(big.Int).Sub(dh.p, two)
	upper := new(big.Int).Sub(dh.p, dhValueMargin)
	for _, v := range []*big.Int{two, pMinusTwo, dhValueMargin, upper} {
		require.Error(t, dh.checkValue(v), "value %s", v)
	}
	inside := new(big.Int).Add(dhValueMargin, bigOne)
	require.NoError(t, dh.checkValue(inside))
	require.NoError(t, dh.checkValue(new(big.Int).Sub(upper, bigOne)))

	// computeKey applies the same range to the peer's value.
	exp := big.NewInt(12345)
	_, _, err := dh.computeKey(pad(two), exp)
	require.Error(t, err)
	_, _, err = dh.computeKey(pad(pMinusTwo), exp)
	require.Error(t, err)
}

// acceptInvoker answers messages.getDhConfig and phone.acceptCall, then
// delivers the caller's phoneCall with the given g_a and key fingerprint.
type acceptInvoker struct {
	t           *testing.T
	discarded   bool
	client      *Client
	gA          []byte
	fingerprint int64
}

func (i *acceptInvoker) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	var result bin.Encoder
	switch input.(type) {
	case *tg.MessagesGetDhConfigRequest:
		result = dhConfigResult(2, testGroup(i.t).p)
	case *tg.PhoneAcceptCallRequest:
		result = &tg.PhonePhoneCall{PhoneCall: &tg.PhoneCallWaiting{ID: 1}}
		go func() {
			_ = i.client.Handle(context.Background(), &tg.UpdatePhoneCall{PhoneCall: &tg.PhoneCall{
				ID:             1,
				GAOrB:          i.gA,
				KeyFingerprint: i.fingerprint,
			}})
		}()
	case *tg.PhoneDiscardCallRequest:
		i.discarded = true
		result = &tg.Updates{}
	default:
		i.t.Fatalf("unexpected request %T", input)
	}
	var b bin.Buffer
	if err := result.Encode(&b); err != nil {
		return err
	}
	return output.Decode(&b)
}

func TestIncomingCallKeyFingerprintMismatch(t *testing.T) {
	dh := testGroup(t)
	_, gAInt, err := dh.randomExp(rand.Reader, nil)
	require.NoError(t, err)
	gA := pad(gAInt)

	inv := &acceptInvoker{t: t, gA: gA, fingerprint: 42}
	c := NewClient(tg.NewClient(inv), Options{})
	inv.client = c

	ic := &IncomingCall{client: c, req: &tg.PhoneCallRequested{
		ID: 1, AccessHash: 2, AdminID: 3, GAHash: gAHash(gA),
	}}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = ic.Accept(ctx)
	require.ErrorContains(t, err, "key fingerprint mismatch")
	require.True(t, inv.discarded, "call must be discarded on the server")

	c.mu.Lock()
	require.Nil(t, c.call, "call must be cleared")
	c.mu.Unlock()
}
