// Binary mtproxy-connect connects to Telegram through an MTProxy and verifies
// that the MTProto connection works, without logging in.
//
// It performs the MTProto handshake through the proxy and calls
// help.getNearestDC, which does not require authorization. This is handy to
// check that a given MTProxy address and secret actually work.
//
// Usage:
//
//	mtproxy-connect "tg://proxy?server=1.2.3.4&port=443&secret=ee<hex...>"
//	mtproxy-connect -addr proxy.example.com:443 -secret ee<hex...>
package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"net"
	"net/url"
	"strings"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
)

// decodeSecret decodes an MTProxy secret that can be either hex-encoded (the
// most common share format) or base64url-encoded (as found in tg://proxy links).
func decodeSecret(s string) ([]byte, error) {
	if b, err := hex.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.URLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return nil, errors.Errorf("unable to decode secret %q as hex or base64url", s)
}

// parseProxyLink parses a proxy share link, e.g.
//
//	tg://proxy?server=1.2.3.4&port=443&secret=ee...
//	https://t.me/proxy?server=1.2.3.4&port=443&secret=ee...
//
// and returns the address (host:port) and the decoded secret.
func parseProxyLink(link string) (addr string, secret []byte, _ error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", nil, errors.Wrap(err, "parse link")
	}

	q := u.Query()
	server, port := q.Get("server"), q.Get("port")
	if server == "" || port == "" {
		return "", nil, errors.New("link is missing server or port")
	}
	secret, err = decodeSecret(q.Get("secret"))
	if err != nil {
		return "", nil, err
	}

	return net.JoinHostPort(server, port), secret, nil
}

func resolveProxy() (addr string, secret []byte, _ error) {
	addrFlag := flag.String("addr", "", "MTProxy address (host:port)")
	secretFlag := flag.String("secret", "", "MTProxy secret (hex or base64url)")
	flag.Parse()

	// A proxy link can be passed as the first positional argument.
	if arg := flag.Arg(0); arg != "" {
		if !strings.HasPrefix(arg, "tg://") && !strings.Contains(arg, "/proxy?") {
			return "", nil, errors.Errorf("unsupported proxy link %q", arg)
		}
		return parseProxyLink(arg)
	}

	if *addrFlag == "" || *secretFlag == "" {
		return "", nil, errors.New("provide a proxy link or both -addr and -secret")
	}
	secret, err := decodeSecret(*secretFlag)
	if err != nil {
		return "", nil, err
	}
	return *addrFlag, secret, nil
}

func main() {
	addr, secret, err := resolveProxy()
	if err != nil {
		flag.Usage()
		panic(err)
	}

	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		resolver, err := dcs.MTProxy(addr, secret, dcs.MTProxyOptions{})
		if err != nil {
			return errors.Wrap(err, "create MTProxy resolver")
		}

		// Using public test credentials: we only check connectivity, so no real
		// application id or authentication is required.
		client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
			Resolver: resolver,
			Logger:   logzap.New(log),
		})

		return client.Run(ctx, func(ctx context.Context) error {
			// help.getNearestDC works without authorization, so a successful
			// response means the MTProto connection through the proxy is healthy.
			dc, err := client.API().HelpGetNearestDC(ctx)
			if err != nil {
				return errors.Wrap(err, "get nearest DC")
			}

			log.Info("Connected to Telegram through MTProxy",
				zap.Int("this_dc", dc.ThisDC),
				zap.Int("nearest_dc", dc.NearestDC),
				zap.String("country", dc.Country),
			)
			return nil
		})
	})
}
