// Package gorules contains ruleguard linter rules.
package gorules

import (
	"github.com/quasilyte/go-ruleguard/dsl"
)

// ZapPreferNoWith suggests replace expressions like
//
// 	l.With(...).Debug("")
//
// to
//
// 	l.Debug("", ...).
//
// where l is a *zap.Logger.
func ZapPreferNoWith(m dsl.Matcher) {
	m.Import("go.uber.org/zap")

	m.Match("$l.With($*args).$method($*msg_args)").Where(
		m["l"].Type.Is("*zap.Logger") &&
			m["method"].Text.Matches("Debug|Info|Warn|Error|DPanic|Panic|Fatal"),
	).Suggest("$l.$method($msg_args, $args)")
}
