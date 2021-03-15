package telegram

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
	const notAvailable = "n/a"

	if c.DeviceModel == "" {
		c.DeviceModel = notAvailable
	}
	if c.SystemVersion == "" {
		c.SystemVersion = notAvailable
	}
	if c.AppVersion == "" {
		c.AppVersion = notAvailable
	}
	if c.SystemLangCode == "" {
		c.SystemLangCode = "en"
	}
	if c.LangCode == "" {
		c.LangCode = "en"
	}
}
