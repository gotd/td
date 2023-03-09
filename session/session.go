// Package session implements session storage.
package session

import (
	"context"
	"encoding/json"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"
)

// Config is subset of tg.Config.
type Config struct {
	// Indicates that telegram is probably censored by governments/ISPs in the current region
	BlockedMode bool
	// Whether to forcefully try connecting using IPv6 dcOptions¹
	//
	// Links:
	//  1) https://core.telegram.org/type/DcOption
	ForceTryIpv6 bool
	// Current date at the server
	Date int
	// Expiration date of this config: when it expires it'll have to be refetched using help
	// getConfig¹
	//
	// Links:
	//  1) https://core.telegram.org/method/help.getConfig
	Expires int
	// Whether we're connected to the test DCs
	TestMode bool
	// ID of the DC that returned the reply
	ThisDC int
	// DC IP list
	DCOptions []tg.DCOption
	// Domain name for fetching encrypted DC list from DNS TXT record
	DCTxtDomainName string
	// Temporary passport¹ sessions
	//
	// Links:
	//  1) https://core.telegram.org/passport
	//
	// Use SetTmpSessions and GetTmpSessions helpers.
	TmpSessions int
	// DC ID to use to download webfiles¹
	//
	// Links:
	//  1) https://core.telegram.org/api/files#downloading-webfiles
	WebfileDCID int
}

// ConfigFromTG converts tg.Config to Config.
//
// Note that Config is the subset of tg.Config, so data loss is possible.
func ConfigFromTG(c tg.Config) Config {
	return Config{
		BlockedMode:     c.BlockedMode,
		ForceTryIpv6:    c.ForceTryIpv6,
		Date:            c.Date,
		Expires:         c.Expires,
		TestMode:        c.TestMode,
		ThisDC:          c.ThisDC,
		DCOptions:       c.DCOptions,
		DCTxtDomainName: c.DCTxtDomainName,
		WebfileDCID:     c.WebfileDCID,
		TmpSessions:     c.TmpSessions,
	}
}

// TG returns tg.Config from Config.
//
// Note that config is the subset of tg.Config, so some fields will be unset.
func (c Config) TG() tg.Config {
	return tg.Config{
		BlockedMode:     c.BlockedMode,
		ForceTryIpv6:    c.ForceTryIpv6,
		Date:            c.Date,
		Expires:         c.Expires,
		TestMode:        c.TestMode,
		ThisDC:          c.ThisDC,
		DCOptions:       c.DCOptions,
		DCTxtDomainName: c.DCTxtDomainName,
		WebfileDCID:     c.WebfileDCID,
		TmpSessions:     c.TmpSessions,
	}
}

// Data of session.
type Data struct {
	Config    Config
	DC        int
	Addr      string
	AuthKey   []byte
	AuthKeyID []byte
	Salt      int64
}

// Storage is secure persistent storage for client session.
//
// NB: Implementation security is important, attacker can abuse it not only for
// connecting as authenticated user or bot, but even decrypting previous
// messages in some situations.
type Storage interface {
	LoadSession(ctx context.Context) ([]byte, error)
	StoreSession(ctx context.Context, data []byte) error
}

// ErrNotFound means that session is not found in storage.
var ErrNotFound = errors.New("session storage: not found")

// Loader wraps Storage implementing Data (un-)marshaling.
type Loader struct {
	Storage Storage
}

type jsonData struct {
	Version int
	Data    Data
}

const latestVersion = 1

// Load loads Data from Storage.
func (l *Loader) Load(ctx context.Context) (*Data, error) {
	buf, err := l.Storage.LoadSession(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "load")
	}
	if len(buf) == 0 {
		return nil, ErrNotFound
	}

	var v jsonData
	if err := json.Unmarshal(buf, &v); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}
	if v.Version != latestVersion {
		// HACK(ernado): backward compatibility super shenanigan.
		return nil, errors.Wrapf(ErrNotFound, "version mismatch (%d != %d)", v.Version, latestVersion)
	}
	return &v.Data, err
}

// Save saves Data to Storage.
func (l *Loader) Save(ctx context.Context, data *Data) error {
	v := jsonData{
		Version: latestVersion,
		Data:    *data,
	}
	buf, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "marshal")
	}
	if err := l.Storage.StoreSession(ctx, buf); err != nil {
		return errors.Wrap(err, "store")
	}
	return nil
}
