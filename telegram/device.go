package telegram

import "runtime"

// DeviceConfig is config which send when Telegram connection session created.
type DeviceConfig struct {
	// Device model
	DeviceModel string
	// Operation system version
	SystemVersion string
	// Application version
	AppVersion string
	// Code for the language used on the device's OS, ISO 639-1 standard
	SystemLangCode string
	// Language pack to use
	LangPack string
	// Code for the language used on the client, ISO 639-1 standard
	LangCode string
}

func (c *DeviceConfig) setDefaults() {
	if c.DeviceModel == "" {
		c.DeviceModel = "gotd"
	}
	if c.SystemVersion == "" {
		c.SystemVersion = runtime.GOOS
	}
	if c.AppVersion == "" {
		c.AppVersion = getVersion()
	}
	if c.SystemLangCode == "" {
		c.SystemLangCode = "en"
	}
	if c.LangPack == "" {
		c.LangPack = ""
	}
	if c.LangCode == "" {
		c.LangCode = "en"
	}
}
