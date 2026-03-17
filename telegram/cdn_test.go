package telegram

import (
	"context"
	"crypto/rsa"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/exchange"
	"github.com/gotd/td/tg"
)

func parseCDNKeysForTest(keys ...tg.CDNPublicKey) ([]*rsa.PublicKey, error) {
	entries, err := parseCDNKeyEntries(keys...)
	if err != nil {
		return nil, err
	}

	r := make([]*rsa.PublicKey, 0, len(entries))
	for _, entry := range entries {
		r = append(r, entry.key)
	}
	return r, nil
}

func Test_parseCDNKeys(t *testing.T) {
	keys := []string{
		`-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA+Lf3PvgE1yxbJUCMaEAk
V0QySTVpnaDjiednB5RbtNWjCeqSVakYHbqqGMIIv5WCGdFdrqOfMNcNSstPtSU6
R9UmRw6tquOIykpSuUOje9H+4XVIKqujyL2ISdK+4ZOMl4hCMkqauw4bP1Sbr03v
ZRQbU6qEA04V4j879BAyBVhr3WG9+Zi+t5XfGSTgSExPYEl8rZNHYNV5RB+BuroV
H2HLTOpT/mJVfikYpgjfWF5ldezV4Wo9LSH0cZGSFIaeJl8d0A8Eiy5B9gtBO8mL
+XfQRKOOmr7a4BM4Ro2de5rr2i2od7hYXd3DO9FRSl4y1zA8Am48Rfd95WHF3N/O
mQIDAQAB
-----END RSA PUBLIC KEY-----`,
		`-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAyu5PXyfp+VFLc2hKJsq/cvQ+wq9V2s1iGMMwcrkXrKAqX0S5QEcY
W9b6pV5LulbsvNcxp/YniiSL4FsAja28B9fH//Y+AolWASomCB0NSVHwS1Pqfe3m
GdLTwDmqU17tSWk/48+Kfn4B+WT85ZIKt8bOnABwnM1AtykX0zKwzm9yKcTX0MeY
rwzgiOQax6J1cfgtLdxl8HVKT6wCOS1e43zpXMU+UoWqRqIan+J6q+ubi1yF4PWl
DyDgJSw8uxlhNNMP4tAnshIRZ1ZZ25O/g58jw1qz5XMztZwLNA2pUxaFtyy1LdHC
FRX7DdwIA/FdOzfWyXYLlCFaSX8K/6CnSQIDAQAB
-----END RSA PUBLIC KEY-----`,
	}

	cdnKeys := make([]tg.CDNPublicKey, 0, len(keys))
	for i, key := range keys {
		cdnKeys = append(cdnKeys, tg.CDNPublicKey{
			DCID:      i + 1,
			PublicKey: key,
		})
	}

	publicKeys, err := parseCDNKeysForTest(cdnKeys...)
	require.NoError(t, err)
	require.Len(t, publicKeys, 2)
}

func Test_fetchCDNKeysRetriesAfterFailure(t *testing.T) {
	// Regression guard:
	// failed first fetch must not poison cache; second call should retry network
	// and then populate cache for subsequent calls.
	a := require.New(t)

	const key = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA+Lf3PvgE1yxbJUCMaEAk
V0QySTVpnaDjiednB5RbtNWjCeqSVakYHbqqGMIIv5WCGdFdrqOfMNcNSstPtSU6
R9UmRw6tquOIykpSuUOje9H+4XVIKqujyL2ISdK+4ZOMl4hCMkqauw4bP1Sbr03v
ZRQbU6qEA04V4j879BAyBVhr3WG9+Zi+t5XfGSTgSExPYEl8rZNHYNV5RB+BuroV
H2HLTOpT/mJVfikYpgjfWF5ldezV4Wo9LSH0cZGSFIaeJl8d0A8Eiy5B9gtBO8mL
+XfQRKOOmr7a4BM4Ro2de5rr2i2od7hYXd3DO9FRSl4y1zA8Am48Rfd95WHF3N/O
mQIDAQAB
-----END RSA PUBLIC KEY-----`

	var calls int
	c := &Client{}
	c.init()
	c.log = zap.NewNop()
	c.tg = tg.NewClient(InvokeFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		_, ok := input.(*tg.HelpGetCDNConfigRequest)
		a.True(ok)
		result, ok := output.(*tg.CDNConfig)
		a.True(ok)

		calls++
		if calls == 1 {
			return errors.New("temporary fetch error")
		}
		result.PublicKeys = []tg.CDNPublicKey{{
			DCID:      1,
			PublicKey: key,
		}}
		return nil
	}))

	_, err := c.fetchCDNKeys(context.Background())
	a.Error(err)

	keys, err := c.fetchCDNKeys(context.Background())
	a.NoError(err)
	a.Len(keys, 1)

	cached, err := c.fetchCDNKeys(context.Background())
	a.NoError(err)
	a.Len(cached, 1)
	a.Equal(2, calls)
}

func Test_fetchCDNKeysInvalidationDropsStaleResult(t *testing.T) {
	// Critical race case:
	// if fingerprint miss invalidates cache while singleflight fetch is in-flight,
	// stale result must be discarded and replaced with fresh key set.
	a := require.New(t)

	const staleKey = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA+Lf3PvgE1yxbJUCMaEAk
V0QySTVpnaDjiednB5RbtNWjCeqSVakYHbqqGMIIv5WCGdFdrqOfMNcNSstPtSU6
R9UmRw6tquOIykpSuUOje9H+4XVIKqujyL2ISdK+4ZOMl4hCMkqauw4bP1Sbr03v
ZRQbU6qEA04V4j879BAyBVhr3WG9+Zi+t5XfGSTgSExPYEl8rZNHYNV5RB+BuroV
H2HLTOpT/mJVfikYpgjfWF5ldezV4Wo9LSH0cZGSFIaeJl8d0A8Eiy5B9gtBO8mL
+XfQRKOOmr7a4BM4Ro2de5rr2i2od7hYXd3DO9FRSl4y1zA8Am48Rfd95WHF3N/O
mQIDAQAB
-----END RSA PUBLIC KEY-----`

	const freshKey = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAyu5PXyfp+VFLc2hKJsq/cvQ+wq9V2s1iGMMwcrkXrKAqX0S5QEcY
W9b6pV5LulbsvNcxp/YniiSL4FsAja28B9fH//Y+AolWASomCB0NSVHwS1Pqfe3m
GdLTwDmqU17tSWk/48+Kfn4B+WT85ZIKt8bOnABwnM1AtykX0zKwzm9yKcTX0MeY
rwzgiOQax6J1cfgtLdxl8HVKT6wCOS1e43zpXMU+UoWqRqIan+J6q+ubi1yF4PWl
DyDgJSw8uxlhNNMP4tAnshIRZ1ZZ25O/g58jw1qz5XMztZwLNA2pUxaFtyy1LdHC
FRX7DdwIA/FdOzfWyXYLlCFaSX8K/6CnSQIDAQAB
-----END RSA PUBLIC KEY-----`

	firstStarted := make(chan struct{})
	unblockFirst := make(chan struct{})

	var calls int
	c := &Client{}
	c.init()
	c.log = zap.NewNop()
	c.tg = tg.NewClient(InvokeFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		_, ok := input.(*tg.HelpGetCDNConfigRequest)
		a.True(ok)
		result, ok := output.(*tg.CDNConfig)
		a.True(ok)

		calls++
		switch calls {
		case 1:
			close(firstStarted)
			<-unblockFirst
			result.PublicKeys = []tg.CDNPublicKey{{
				DCID:      1,
				PublicKey: staleKey,
			}}
		default:
			result.PublicKeys = []tg.CDNPublicKey{{
				DCID:      1,
				PublicKey: freshKey,
			}}
		}

		return nil
	}))

	type fetchResult struct {
		keys []exchange.PublicKey
		err  error
	}
	done := make(chan fetchResult, 1)
	go func() {
		keys, err := c.fetchCDNKeys(context.Background())
		done <- fetchResult{keys: keys, err: err}
	}()

	<-firstStarted
	c.handleCDNConnDead(203, exchange.ErrKeyFingerprintNotFound)
	close(unblockFirst)

	result := <-done
	a.NoError(result.err)
	a.Len(result.keys, 1)
	a.Equal(2, calls)

	parsed, err := parseCDNKeysForTest(tg.CDNPublicKey{DCID: 1, PublicKey: freshKey})
	a.NoError(err)
	a.Len(parsed, 1)
	a.Equal(exchange.PublicKey{RSA: parsed[0]}.Fingerprint(), result.keys[0].Fingerprint())

	cached, err := c.fetchCDNKeys(context.Background())
	a.NoError(err)
	a.Len(cached, 1)
	a.Equal(result.keys[0].Fingerprint(), cached[0].Fingerprint())
	a.Equal(2, calls)
}

func Test_refreshCDNKeysInvalidationDropsStaleResult(t *testing.T) {
	// Critical race case for forced refresh path:
	// if fingerprint miss happens while refresh is in-flight, stale result must
	// not become the cached keyset.
	a := require.New(t)

	const staleKey = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA+Lf3PvgE1yxbJUCMaEAk
V0QySTVpnaDjiednB5RbtNWjCeqSVakYHbqqGMIIv5WCGdFdrqOfMNcNSstPtSU6
R9UmRw6tquOIykpSuUOje9H+4XVIKqujyL2ISdK+4ZOMl4hCMkqauw4bP1Sbr03v
ZRQbU6qEA04V4j879BAyBVhr3WG9+Zi+t5XfGSTgSExPYEl8rZNHYNV5RB+BuroV
H2HLTOpT/mJVfikYpgjfWF5ldezV4Wo9LSH0cZGSFIaeJl8d0A8Eiy5B9gtBO8mL
+XfQRKOOmr7a4BM4Ro2de5rr2i2od7hYXd3DO9FRSl4y1zA8Am48Rfd95WHF3N/O
mQIDAQAB
-----END RSA PUBLIC KEY-----`

	const freshKey = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAyu5PXyfp+VFLc2hKJsq/cvQ+wq9V2s1iGMMwcrkXrKAqX0S5QEcY
W9b6pV5LulbsvNcxp/YniiSL4FsAja28B9fH//Y+AolWASomCB0NSVHwS1Pqfe3m
GdLTwDmqU17tSWk/48+Kfn4B+WT85ZIKt8bOnABwnM1AtykX0zKwzm9yKcTX0MeY
rwzgiOQax6J1cfgtLdxl8HVKT6wCOS1e43zpXMU+UoWqRqIan+J6q+ubi1yF4PWl
DyDgJSw8uxlhNNMP4tAnshIRZ1ZZ25O/g58jw1qz5XMztZwLNA2pUxaFtyy1LdHC
FRX7DdwIA/FdOzfWyXYLlCFaSX8K/6CnSQIDAQAB
-----END RSA PUBLIC KEY-----`

	firstStarted := make(chan struct{})
	unblockFirst := make(chan struct{})
	var calls atomic.Int32

	c := &Client{}
	c.init()
	c.log = zap.NewNop()
	c.tg = tg.NewClient(InvokeFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		_, ok := input.(*tg.HelpGetCDNConfigRequest)
		a.True(ok)
		result, ok := output.(*tg.CDNConfig)
		a.True(ok)

		switch calls.Add(1) {
		case 1:
			close(firstStarted)
			<-unblockFirst
			result.PublicKeys = []tg.CDNPublicKey{{
				DCID:      1,
				PublicKey: staleKey,
			}}
		default:
			result.PublicKeys = []tg.CDNPublicKey{{
				DCID:      1,
				PublicKey: freshKey,
			}}
		}

		return nil
	}))

	type fetchResult struct {
		keys []exchange.PublicKey
		err  error
	}
	done := make(chan fetchResult, 1)
	go func() {
		keys, err := c.refreshCDNKeys(context.Background())
		done <- fetchResult{keys: keys, err: err}
	}()

	<-firstStarted
	c.handleCDNConnDead(203, exchange.ErrKeyFingerprintNotFound)
	close(unblockFirst)

	refreshResult := <-done
	a.NoError(refreshResult.err)
	a.Len(refreshResult.keys, 1)

	parsedFresh, err := parseCDNKeysForTest(tg.CDNPublicKey{
		DCID:      1,
		PublicKey: freshKey,
	})
	a.NoError(err)
	a.Len(parsedFresh, 1)
	freshFingerprint := exchange.PublicKey{RSA: parsedFresh[0]}.Fingerprint()

	keys, err := c.fetchCDNKeys(context.Background())
	a.NoError(err)
	a.Len(keys, 1)
	a.Equal(freshFingerprint, keys[0].Fingerprint())
	a.GreaterOrEqual(calls.Load(), int32(2))

	callsBeforeCacheRead := calls.Load()
	cached, err := c.fetchCDNKeys(context.Background())
	a.NoError(err)
	a.Len(cached, 1)
	a.Equal(freshFingerprint, cached[0].Fingerprint())
	a.Equal(callsBeforeCacheRead, calls.Load(), "second fetch should hit cache")
}

func Test_fetchCDNKeysForDCReturnsOnlyRequestedDC(t *testing.T) {
	a := require.New(t)

	const keyDC1 = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA+Lf3PvgE1yxbJUCMaEAk
V0QySTVpnaDjiednB5RbtNWjCeqSVakYHbqqGMIIv5WCGdFdrqOfMNcNSstPtSU6
R9UmRw6tquOIykpSuUOje9H+4XVIKqujyL2ISdK+4ZOMl4hCMkqauw4bP1Sbr03v
ZRQbU6qEA04V4j879BAyBVhr3WG9+Zi+t5XfGSTgSExPYEl8rZNHYNV5RB+BuroV
H2HLTOpT/mJVfikYpgjfWF5ldezV4Wo9LSH0cZGSFIaeJl8d0A8Eiy5B9gtBO8mL
+XfQRKOOmr7a4BM4Ro2de5rr2i2od7hYXd3DO9FRSl4y1zA8Am48Rfd95WHF3N/O
mQIDAQAB
-----END RSA PUBLIC KEY-----`
	const keyDC2 = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAyu5PXyfp+VFLc2hKJsq/cvQ+wq9V2s1iGMMwcrkXrKAqX0S5QEcY
W9b6pV5LulbsvNcxp/YniiSL4FsAja28B9fH//Y+AolWASomCB0NSVHwS1Pqfe3m
GdLTwDmqU17tSWk/48+Kfn4B+WT85ZIKt8bOnABwnM1AtykX0zKwzm9yKcTX0MeY
rwzgiOQax6J1cfgtLdxl8HVKT6wCOS1e43zpXMU+UoWqRqIan+J6q+ubi1yF4PWl
DyDgJSw8uxlhNNMP4tAnshIRZ1ZZ25O/g58jw1qz5XMztZwLNA2pUxaFtyy1LdHC
FRX7DdwIA/FdOzfWyXYLlCFaSX8K/6CnSQIDAQAB
-----END RSA PUBLIC KEY-----`

	var calls int
	c := &Client{}
	c.init()
	c.log = zap.NewNop()
	c.tg = tg.NewClient(InvokeFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		_, ok := input.(*tg.HelpGetCDNConfigRequest)
		a.True(ok)
		result, ok := output.(*tg.CDNConfig)
		a.True(ok)
		calls++
		result.PublicKeys = []tg.CDNPublicKey{
			{
				DCID:      1,
				PublicKey: keyDC1,
			},
			{
				DCID:      2,
				PublicKey: keyDC2,
			},
		}
		return nil
	}))

	all, err := c.fetchCDNKeys(context.Background())
	a.NoError(err)
	a.Len(all, 2)

	dc1, err := c.fetchCDNKeysForDC(context.Background(), 1)
	a.NoError(err)
	a.Len(dc1, 1)

	dc2, err := c.fetchCDNKeysForDC(context.Background(), 2)
	a.NoError(err)
	a.Len(dc2, 1)

	a.NotEqual(dc1[0].Fingerprint(), dc2[0].Fingerprint())
	a.Equal(1, calls, "help.getCdnConfig must stay cached")
}

func Test_fetchCDNKeysCanceledCallerDoesNotPoisonConcurrentWaiters(t *testing.T) {
	a := require.New(t)

	const key = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA+Lf3PvgE1yxbJUCMaEAk
V0QySTVpnaDjiednB5RbtNWjCeqSVakYHbqqGMIIv5WCGdFdrqOfMNcNSstPtSU6
R9UmRw6tquOIykpSuUOje9H+4XVIKqujyL2ISdK+4ZOMl4hCMkqauw4bP1Sbr03v
ZRQbU6qEA04V4j879BAyBVhr3WG9+Zi+t5XfGSTgSExPYEl8rZNHYNV5RB+BuroV
H2HLTOpT/mJVfikYpgjfWF5ldezV4Wo9LSH0cZGSFIaeJl8d0A8Eiy5B9gtBO8mL
+XfQRKOOmr7a4BM4Ro2de5rr2i2od7hYXd3DO9FRSl4y1zA8Am48Rfd95WHF3N/O
mQIDAQAB
-----END RSA PUBLIC KEY-----`

	started := make(chan struct{})
	release := make(chan struct{})
	var startedOnce sync.Once
	var calls atomic.Int32

	c := &Client{}
	c.init()
	c.log = zap.NewNop()
	c.ctx, c.cancel = context.WithCancel(context.Background())
	defer c.cancel()
	c.tg = tg.NewClient(InvokeFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		_, ok := input.(*tg.HelpGetCDNConfigRequest)
		a.True(ok)
		result, ok := output.(*tg.CDNConfig)
		a.True(ok)

		calls.Add(1)
		startedOnce.Do(func() {
			close(started)
		})

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-release:
			result.PublicKeys = []tg.CDNPublicKey{{
				DCID:      1,
				PublicKey: key,
			}}
			return nil
		}
	}))

	type fetchResult struct {
		keys []exchange.PublicKey
		err  error
	}

	firstCtx, firstCancel := context.WithCancel(context.Background())
	defer firstCancel()

	firstDone := make(chan fetchResult, 1)
	go func() {
		keys, err := c.fetchCDNKeys(firstCtx)
		firstDone <- fetchResult{keys: keys, err: err}
	}()

	<-started

	secondDone := make(chan fetchResult, 1)
	go func() {
		keys, err := c.fetchCDNKeys(context.Background())
		secondDone <- fetchResult{keys: keys, err: err}
	}()

	firstCancel()
	close(release)

	first := <-firstDone
	second := <-secondDone

	a.ErrorIs(first.err, context.Canceled)
	a.NoError(second.err)
	a.Len(second.keys, 1)
	a.GreaterOrEqual(calls.Load(), int32(1))
	a.LessOrEqual(calls.Load(), int32(2))
}

func Test_fetchCDNKeysDeadlineCallerDoesNotPoisonConcurrentWaiters(t *testing.T) {
	a := require.New(t)

	const key = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA+Lf3PvgE1yxbJUCMaEAk
V0QySTVpnaDjiednB5RbtNWjCeqSVakYHbqqGMIIv5WCGdFdrqOfMNcNSstPtSU6
R9UmRw6tquOIykpSuUOje9H+4XVIKqujyL2ISdK+4ZOMl4hCMkqauw4bP1Sbr03v
ZRQbU6qEA04V4j879BAyBVhr3WG9+Zi+t5XfGSTgSExPYEl8rZNHYNV5RB+BuroV
H2HLTOpT/mJVfikYpgjfWF5ldezV4Wo9LSH0cZGSFIaeJl8d0A8Eiy5B9gtBO8mL
+XfQRKOOmr7a4BM4Ro2de5rr2i2od7hYXd3DO9FRSl4y1zA8Am48Rfd95WHF3N/O
mQIDAQAB
-----END RSA PUBLIC KEY-----`

	started := make(chan struct{})
	release := make(chan struct{})
	var startedOnce sync.Once
	var calls atomic.Int32

	c := &Client{}
	c.init()
	c.log = zap.NewNop()
	c.ctx, c.cancel = context.WithCancel(context.Background())
	defer c.cancel()
	c.tg = tg.NewClient(InvokeFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		_, ok := input.(*tg.HelpGetCDNConfigRequest)
		a.True(ok)
		result, ok := output.(*tg.CDNConfig)
		a.True(ok)

		calls.Add(1)
		startedOnce.Do(func() {
			close(started)
		})

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-release:
			result.PublicKeys = []tg.CDNPublicKey{{
				DCID:      1,
				PublicKey: key,
			}}
			return nil
		}
	}))

	type fetchResult struct {
		keys []exchange.PublicKey
		err  error
	}

	firstCtx, firstCancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer firstCancel()

	firstDone := make(chan fetchResult, 1)
	go func() {
		keys, err := c.fetchCDNKeys(firstCtx)
		firstDone <- fetchResult{keys: keys, err: err}
	}()

	<-started

	secondDone := make(chan fetchResult, 1)
	go func() {
		keys, err := c.fetchCDNKeys(context.Background())
		secondDone <- fetchResult{keys: keys, err: err}
	}()

	time.Sleep(35 * time.Millisecond)
	close(release)

	first := <-firstDone
	second := <-secondDone

	a.Error(first.err)
	a.True(errors.Is(first.err, context.DeadlineExceeded))
	a.NoError(second.err)
	a.Len(second.keys, 1)
	a.GreaterOrEqual(calls.Load(), int32(1))
	a.LessOrEqual(calls.Load(), int32(2))
}

func Test_fetchCDNKeysForDCRetriesWhenCachedConfigMissesRequestedDC(t *testing.T) {
	a := require.New(t)

	const keyDC1 = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA+Lf3PvgE1yxbJUCMaEAk
V0QySTVpnaDjiednB5RbtNWjCeqSVakYHbqqGMIIv5WCGdFdrqOfMNcNSstPtSU6
R9UmRw6tquOIykpSuUOje9H+4XVIKqujyL2ISdK+4ZOMl4hCMkqauw4bP1Sbr03v
ZRQbU6qEA04V4j879BAyBVhr3WG9+Zi+t5XfGSTgSExPYEl8rZNHYNV5RB+BuroV
H2HLTOpT/mJVfikYpgjfWF5ldezV4Wo9LSH0cZGSFIaeJl8d0A8Eiy5B9gtBO8mL
+XfQRKOOmr7a4BM4Ro2de5rr2i2od7hYXd3DO9FRSl4y1zA8Am48Rfd95WHF3N/O
mQIDAQAB
-----END RSA PUBLIC KEY-----`
	const keyDC2 = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAyu5PXyfp+VFLc2hKJsq/cvQ+wq9V2s1iGMMwcrkXrKAqX0S5QEcY
W9b6pV5LulbsvNcxp/YniiSL4FsAja28B9fH//Y+AolWASomCB0NSVHwS1Pqfe3m
GdLTwDmqU17tSWk/48+Kfn4B+WT85ZIKt8bOnABwnM1AtykX0zKwzm9yKcTX0MeY
rwzgiOQax6J1cfgtLdxl8HVKT6wCOS1e43zpXMU+UoWqRqIan+J6q+ubi1yF4PWl
DyDgJSw8uxlhNNMP4tAnshIRZ1ZZ25O/g58jw1qz5XMztZwLNA2pUxaFtyy1LdHC
FRX7DdwIA/FdOzfWyXYLlCFaSX8K/6CnSQIDAQAB
-----END RSA PUBLIC KEY-----`

	var calls int
	c := &Client{}
	c.init()
	c.log = zap.NewNop()
	c.tg = tg.NewClient(InvokeFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		_, ok := input.(*tg.HelpGetCDNConfigRequest)
		a.True(ok)
		result, ok := output.(*tg.CDNConfig)
		a.True(ok)

		calls++
		switch calls {
		case 1:
			// First load has no keys for DC 2.
			result.PublicKeys = []tg.CDNPublicKey{{
				DCID:      1,
				PublicKey: keyDC1,
			}}
		default:
			// Refresh includes keys for requested DC.
			result.PublicKeys = []tg.CDNPublicKey{
				{
					DCID:      1,
					PublicKey: keyDC1,
				},
				{
					DCID:      2,
					PublicKey: keyDC2,
				},
			}
		}
		return nil
	}))

	_, err := c.fetchCDNKeys(context.Background())
	a.NoError(err)
	a.Equal(1, calls)

	keysDC2, err := c.fetchCDNKeysForDC(context.Background(), 2)
	a.NoError(err)
	a.Len(keysDC2, 1)
	a.Equal(2, calls)

	keysDC2Cached, err := c.fetchCDNKeysForDC(context.Background(), 2)
	a.NoError(err)
	a.Len(keysDC2Cached, 1)
	a.Equal(2, calls, "successful refresh should be cached")
}

func Test_fetchCDNKeysForDCRecoversFromEmptyCachedSnapshot(t *testing.T) {
	a := require.New(t)

	const keyDC2 = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAyu5PXyfp+VFLc2hKJsq/cvQ+wq9V2s1iGMMwcrkXrKAqX0S5QEcY
W9b6pV5LulbsvNcxp/YniiSL4FsAja28B9fH//Y+AolWASomCB0NSVHwS1Pqfe3m
GdLTwDmqU17tSWk/48+Kfn4B+WT85ZIKt8bOnABwnM1AtykX0zKwzm9yKcTX0MeY
rwzgiOQax6J1cfgtLdxl8HVKT6wCOS1e43zpXMU+UoWqRqIan+J6q+ubi1yF4PWl
DyDgJSw8uxlhNNMP4tAnshIRZ1ZZ25O/g58jw1qz5XMztZwLNA2pUxaFtyy1LdHC
FRX7DdwIA/FdOzfWyXYLlCFaSX8K/6CnSQIDAQAB
-----END RSA PUBLIC KEY-----`

	var calls int
	c := &Client{}
	c.init()
	c.log = zap.NewNop()
	c.cdnKeysSet = true
	c.cdnKeys = nil
	c.cdnKeysByDC = map[int][]PublicKey{}
	c.tg = tg.NewClient(InvokeFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		_, ok := input.(*tg.HelpGetCDNConfigRequest)
		a.True(ok)
		cfg, ok := output.(*tg.CDNConfig)
		a.True(ok)

		calls++
		cfg.PublicKeys = []tg.CDNPublicKey{{
			DCID:      2,
			PublicKey: keyDC2,
		}}
		return nil
	}))

	keys, err := c.fetchCDNKeysForDC(context.Background(), 2)
	a.NoError(err)
	a.Len(keys, 1)
	a.Equal(1, calls)

	// Ensure recovered keyset is now cached and reused.
	keysCached, err := c.fetchCDNKeysForDC(context.Background(), 2)
	a.NoError(err)
	a.Len(keysCached, 1)
	a.Equal(1, calls)
}

func Test_fetchCDNKeysForDCMissingAfterRefreshRecoversWithinSingleCall(t *testing.T) {
	a := require.New(t)

	const keyDC1 = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA+Lf3PvgE1yxbJUCMaEAk
V0QySTVpnaDjiednB5RbtNWjCeqSVakYHbqqGMIIv5WCGdFdrqOfMNcNSstPtSU6
R9UmRw6tquOIykpSuUOje9H+4XVIKqujyL2ISdK+4ZOMl4hCMkqauw4bP1Sbr03v
ZRQbU6qEA04V4j879BAyBVhr3WG9+Zi+t5XfGSTgSExPYEl8rZNHYNV5RB+BuroV
H2HLTOpT/mJVfikYpgjfWF5ldezV4Wo9LSH0cZGSFIaeJl8d0A8Eiy5B9gtBO8mL
+XfQRKOOmr7a4BM4Ro2de5rr2i2od7hYXd3DO9FRSl4y1zA8Am48Rfd95WHF3N/O
mQIDAQAB
-----END RSA PUBLIC KEY-----`
	const keyDC2 = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAyu5PXyfp+VFLc2hKJsq/cvQ+wq9V2s1iGMMwcrkXrKAqX0S5QEcY
W9b6pV5LulbsvNcxp/YniiSL4FsAja28B9fH//Y+AolWASomCB0NSVHwS1Pqfe3m
GdLTwDmqU17tSWk/48+Kfn4B+WT85ZIKt8bOnABwnM1AtykX0zKwzm9yKcTX0MeY
rwzgiOQax6J1cfgtLdxl8HVKT6wCOS1e43zpXMU+UoWqRqIan+J6q+ubi1yF4PWl
DyDgJSw8uxlhNNMP4tAnshIRZ1ZZ25O/g58jw1qz5XMztZwLNA2pUxaFtyy1LdHC
FRX7DdwIA/FdOzfWyXYLlCFaSX8K/6CnSQIDAQAB
-----END RSA PUBLIC KEY-----`

	var calls int
	c := &Client{}
	c.init()
	c.log = zap.NewNop()
	c.tg = tg.NewClient(InvokeFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		_, ok := input.(*tg.HelpGetCDNConfigRequest)
		a.True(ok)
		result, ok := output.(*tg.CDNConfig)
		a.True(ok)

		calls++
		switch calls {
		case 1:
			// Initial load has no keys for DC 2.
			result.PublicKeys = []tg.CDNPublicKey{{
				DCID:      1,
				PublicKey: keyDC1,
			}}
		case 2:
			// First refresh still misses requested DC.
			result.PublicKeys = []tg.CDNPublicKey{{
				DCID:      1,
				PublicKey: keyDC1,
			}}
		default:
			// Next refresh returns keys for requested DC.
			result.PublicKeys = []tg.CDNPublicKey{
				{
					DCID:      1,
					PublicKey: keyDC1,
				},
				{
					DCID:      2,
					PublicKey: keyDC2,
				},
			}
		}
		return nil
	}))

	_, err := c.fetchCDNKeys(context.Background())
	a.NoError(err)
	a.Equal(1, calls)

	recovered, err := c.fetchCDNKeysForDC(context.Background(), 2)
	a.NoError(err)
	a.Len(recovered, 1)
	a.Equal(3, calls)

	cached, err := c.fetchCDNKeysForDC(context.Background(), 2)
	a.NoError(err)
	a.Len(cached, 1)
	a.Equal(3, calls, "successful recovery should be cached")
}
