package telegram

import (
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"
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

	// Dialer to use. Default dialer will be used if not provided.
	Dialer Dialer
	// Network to use. Defaults to tcp.
	Network string
	// Ping duration. Default 1 minute.
	PingDuration time.Duration
	// Ping timeout. Default 15 seconds.
	PingTimeout time.Duration
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
	if opt.Dialer == nil {
		opt.Dialer = &net.Dialer{}
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
	if opt.PingDuration.Nanoseconds() <= 0 {
		opt.PingDuration = time.Minute
	}
	if opt.PingTimeout.Nanoseconds() <= 0 {
		opt.PingTimeout = time.Second * 15
	}
}
