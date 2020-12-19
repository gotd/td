package telegram

import (
	"crypto/rand"
	"crypto/rsa"
	"io"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/transport"
)

// Options of Client.
type Options struct {
	// PublicKeys of telegram.
	//
	// If not provided, embedded public keys will be used.
	PublicKeys []*rsa.PublicKey

	// Addr to connect.
	//
	// If not provided, AddrProduction will be used by default.
	Addr string

	// Transport to use. Default dialer will be used if not provided.
	Transport *transport.Transport
	// Network to use. Defaults to tcp.
	Network string
	// Random is random source. Defaults to crypto.
	Random io.Reader
	// Logger is instance of zap.Logger. No logs by default.
	Logger *zap.Logger
	// SessionStorage will be used to load and save session data.
	// NB: Very sensitive data, save with care.
	SessionStorage SessionStorage
	// UpdateHandler will be called on received update.
	UpdateHandler UpdateHandler
}

func (opt *Options) setDefaults() {
	if opt.Transport == nil {
		opt.Transport = transport.Intermediate(nil)
	}
	if opt.Network == "" {
		opt.Network = "tcp"
	}
	if opt.Random == nil {
		opt.Random = rand.Reader
	}
	if opt.Logger == nil {
		opt.Logger = zap.NewNop()
	}
	if opt.Addr == "" {
		opt.Addr = AddrProduction
	}
	if len(opt.PublicKeys) == 0 {
		// Using public keys that are included with distribution if not
		// provided.
		//
		// This should never fail and keys should be valid for recent
		// library versions.
		keys, err := vendoredKeys()
		if err != nil {
			panic(xerrors.Errorf("load vendored keys: %w", err))
		}
		opt.PublicKeys = keys
	}
}
