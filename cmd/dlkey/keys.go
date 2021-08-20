package main

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"net/url"
	"sort"

	"go.uber.org/multierr"
	"golang.org/x/xerrors"

	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/keyparser"
)

func get(ctx context.Context, u string) (_ io.ReadCloser, rErr error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, xerrors.Errorf("create request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, xerrors.Errorf("send request: %w", err)
	}
	defer func() {
		if rErr != nil && res != nil && res.Body != nil {
			rErr = multierr.Append(rErr, res.Body.Close())
		}
	}()
	if res.StatusCode/100 != 2 {
		return nil, xerrors.Errorf("unexpected code %d", res.StatusCode)
	}

	return res.Body, nil
}

type telegramKey struct {
	Public      *rsa.PublicKey
	Fingerprint int64
}

type telegramKeys []telegramKey

func (k telegramKeys) Print(w io.Writer) error {
	sort.Stable(k)

	for _, key := range k {
		if err := pem.Encode(w, &pem.Block{
			Type:    "PUBLIC KEY",
			Headers: nil,
			Bytes:   x509.MarshalPKCS1PublicKey(key.Public),
		}); err != nil {
			return xerrors.Errorf("encode key %#x: %w", key.Fingerprint, err)
		}
	}

	return nil
}

func (k *telegramKeys) Add(key *rsa.PublicKey) {
	*k = append(*k, telegramKey{
		Public:      key,
		Fingerprint: crypto.RSAFingerprint(key),
	})
}

func (k telegramKeys) Find(f int64) (telegramKey, bool) {
	l := k.Len()
	idx := sort.Search(l, func(idx int) bool {
		return k[idx].Fingerprint <= f
	})

	if idx < 0 || idx >= l {
		return telegramKey{}, false
	}

	return k[idx], true
}

func (k telegramKeys) Len() int {
	return len(k)
}

func (k telegramKeys) Less(i, j int) bool {
	return k[i].Fingerprint < k[j].Fingerprint
}

func (k telegramKeys) Swap(i, j int) {
	k[i], k[j] = k[j], k[i]
}

func extractKeys(ctx context.Context, u *url.URL) (_ telegramKeys, rErr error) {
	res, err := get(ctx, u.String())
	if err != nil {
		return nil, xerrors.Errorf("get: %w", err)
	}
	defer multierr.AppendInvoke(&rErr, multierr.Close(res))

	buf := bytes.Buffer{}
	if err := keyparser.Extract(res, &buf); err != nil {
		return nil, xerrors.Errorf("extract: %w", err)
	}

	keys, err := crypto.ParseRSAPublicKeys(buf.Bytes())
	if err != nil {
		return nil, xerrors.Errorf("parse: %w", err)
	}

	r := make(telegramKeys, 0, len(keys))
	for _, key := range keys {
		r.Add(key)
	}
	sort.Stable(r)

	return r, nil
}
