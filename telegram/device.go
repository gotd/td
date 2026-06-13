package telegram

import (
	"time"

	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/internal/manager"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/transport"
)

// DeviceConfig is config which send when Telegram connection session created.
type DeviceConfig = manager.DeviceConfig

// Telegram Desktop application version sent in initConnection.
//
// Keep in sync with the bundled tdesktop reference (Telegram/SourceFiles/core/version.h).
const tdesktopAppVersion = "6.9.1"

// TimezoneParams builds the initConnection params value with the tz_offset
// field, matching what the official clients (e.g. Telegram Desktop) send.
//
// The offset is the timezone offset of loc in seconds, rounded to the nearest
// 15 minutes (900 seconds) and clamped to the [-12h, +14h] range, exactly like
// Telegram Desktop's prepareInitParams.
func TimezoneParams(loc *time.Location) tg.JSONValueClass {
	_, offset := time.Now().In(loc).Zone()

	for offset < -12*3600 {
		offset += 24 * 3600
	}
	for offset > 14*3600 {
		offset -= 24 * 3600
	}

	sign := 1
	if offset < 0 {
		sign = -1
		offset = -offset
	}
	rounded := ((offset + 450) / 900) * 900 * sign

	return &tg.JSONObject{
		Value: []tg.JSONObjectValue{
			{Key: "tz_offset", Value: &tg.JSONNumber{Value: float64(rounded)}},
		},
	}
}

// TDesktopResolver returns a DC resolver that connects the same way as Telegram
// Desktop: every direct connection is wrapped in Obfuscated2 and uses the
// abridged transport codec.
//
// Use it together with DeviceTDesktopWindows to be indistinguishable from
// Telegram Desktop both on the transport layer and in initConnection:
//
//	client := telegram.NewClient(appID, appHash, telegram.Options{
//		Device:   telegram.DeviceTDesktopWindows(),
//		Resolver: telegram.TDesktopResolver(),
//	})
func TDesktopResolver() dcs.Resolver {
	return dcs.Plain(dcs.PlainOptions{
		Protocol:   transport.Abridged,
		Obfuscated: true,
	})
}

// DeviceTDesktopWindows returns a DeviceConfig that emulates a Telegram Desktop
// (Windows build) installation on initConnection, making the connection
// parameters indistinguishable from Telegram Desktop from the server's
// perspective.
//
// Combine it with TDesktopResolver to also match Telegram Desktop on the
// transport layer.
func DeviceTDesktopWindows() DeviceConfig {
	return DeviceConfig{
		DeviceModel:    "Desktop",
		SystemVersion:  "Windows 10",
		AppVersion:     tdesktopAppVersion + " x64",
		SystemLangCode: "en-US",
		LangPack:       "tdesktop",
		LangCode:       "en",
		Params:         TimezoneParams(time.Local),
	}
}
