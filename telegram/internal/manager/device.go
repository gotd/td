package manager

import (
	"runtime"

	"github.com/gotd/td/telegram/internal/version"
	"github.com/gotd/td/tg"
)

// DeviceConfig is config which send when Telegram connection session created.
type DeviceConfig struct {
	// Device model.
	DeviceModel string
	// Operating system version.
	SystemVersion string
	// Application version.
	AppVersion string
	// Code for the language used on the device's OS, ISO 639-1 standard.
	SystemLangCode string
	// Language pack to use.
	LangPack string
	// Code for the language used on the client, ISO 639-1 standard.
	LangCode string
	// Info about an MTProto proxy.
	Proxy tg.InputClientProxy
	// Additional initConnection parameters. For now, only the tz_offset field is supported,
	// for specifying timezone offset in seconds.
	Params tg.JSONValueClass
}

// SetDefaults sets default values.
func (c *DeviceConfig) SetDefaults() {
	const notAvailable = "n/a"

	// Strings must be non-empty, so set notAvailable if default value is empty.
	set := func(to *string, value string) {
		if value != "" {
			*to = value
		} else {
			*to = notAvailable
		}
	}

	if c.DeviceModel == "" {
		set(&c.DeviceModel, runtime.Version())
	}
	if c.SystemVersion == "" {
		set(&c.SystemVersion, runtime.GOOS)
	}
	if c.AppVersion == "" {
		set(&c.AppVersion, version.GetVersion())
	}
	if c.SystemLangCode == "" {
		c.SystemLangCode = "en"
	}
	if c.LangCode == "" {
		c.LangCode = "en"
	}
	// It's okay to use zero value Proxy.
	// It's okay to use zero value Params.
}
