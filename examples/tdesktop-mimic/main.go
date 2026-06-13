// Binary tdesktop-mimic connects to Telegram so that the connection is
// indistinguishable from the official Telegram Desktop client from the server's
// perspective, without logging in.
//
// It combines two presets:
//
//   - telegram.TDesktopResolver: every direct connection is wrapped in
//     Obfuscated2 and uses the abridged transport codec, exactly like Telegram
//     Desktop (which always obfuscates direct connections).
//   - telegram.DeviceTDesktopWindows: initConnection reports a Telegram Desktop
//     (Windows) installation — device model, system/app version, lang_pack
//     "tdesktop" and a tz_offset param.
//
// It then calls help.getNearestDC, which does not require authorization, so a
// successful response means the mimicking MTProto connection is healthy.
//
// Usage:
//
//	tdesktop-mimic
package main

import (
	"context"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/log/logzap"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/telegram"
)

func main() {
	examples.Run(func(ctx context.Context, log *zap.Logger) error {
		// Using public test credentials: we only check connectivity, so no real
		// application id or authentication is required.
		client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
			// Match Telegram Desktop on the transport layer (Obfuscated2 + abridged).
			Resolver: telegram.TDesktopResolver(),
			// Match Telegram Desktop in initConnection (device, versions, lang_pack, tz_offset).
			Device: telegram.DeviceTDesktopWindows(),
			Logger: logzap.New(log),
		})

		return client.Run(ctx, func(ctx context.Context) error {
			// help.getNearestDC works without authorization, so a successful
			// response means the mimicking MTProto connection is healthy.
			dc, err := client.API().HelpGetNearestDC(ctx)
			if err != nil {
				return errors.Wrap(err, "get nearest DC")
			}

			device := telegram.DeviceTDesktopWindows()
			log.Info("Connected to Telegram mimicking Telegram Desktop",
				zap.Int("this_dc", dc.ThisDC),
				zap.Int("nearest_dc", dc.NearestDC),
				zap.String("country", dc.Country),
				zap.String("device_model", device.DeviceModel),
				zap.String("system_version", device.SystemVersion),
				zap.String("app_version", device.AppVersion),
				zap.String("lang_pack", device.LangPack),
			)
			return nil
		})
	})
}
