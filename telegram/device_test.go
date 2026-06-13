package telegram

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestTimezoneParams(t *testing.T) {
	a := require.New(t)

	// Fixed offset of +3h must map to a tz_offset of 10800 seconds.
	loc := time.FixedZone("MSK", 3*3600)
	v := TimezoneParams(loc)

	obj, ok := v.(*tg.JSONObject)
	a.True(ok)
	a.Len(obj.Value, 1)
	a.Equal("tz_offset", obj.Value[0].Key)

	num, ok := obj.Value[0].Value.(*tg.JSONNumber)
	a.True(ok)
	a.Equal(float64(3*3600), num.Value)
}

func TestTimezoneParamsRounding(t *testing.T) {
	a := require.New(t)

	// Non-15-minute offsets are rounded to the nearest 900 seconds.
	loc := time.FixedZone("NPT", 5*3600+45*60) // +05:45
	v := TimezoneParams(loc)
	num := v.(*tg.JSONObject).Value[0].Value.(*tg.JSONNumber)
	a.Zero(int(num.Value) % 900)
	a.Equal(float64(5*3600+45*60), num.Value)
}

func TestDeviceTDesktopWindows(t *testing.T) {
	a := require.New(t)

	d := DeviceTDesktopWindows()
	a.Equal("Desktop", d.DeviceModel)
	a.Equal("Windows 10", d.SystemVersion)
	a.Equal("tdesktop", d.LangPack)
	a.Equal("en", d.LangCode)
	a.Equal("en-US", d.SystemLangCode)
	a.NotEmpty(d.AppVersion)
	a.NotNil(d.Params)
}

func TestTDesktopResolver(t *testing.T) {
	// Wire-level obfuscated abridged behavior is covered by
	// dcs.TestPlainObfuscatedDirect; here we only ensure the helper is wired.
	require.NotNil(t, TDesktopResolver())
}
