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

// UberStyleErrors detects errors messages like
//
// 	xerrors.Errorf("failed to do something: %w", err)
//
// to
//
// 	xerrors.Errorf("do something: %w", err)
//
// according to https://github.com/uber-go/guide/blob/master/style.md#error-wrapping.
func UberStyleErrors(m dsl.Matcher) {
	m.Match("$pkg.Errorf($msg, $*msg_args)").Where(
		m["msg"].Text.Matches(`"failed to.*"`),
	).Report("Avoid phrases like \"failed to\"")
}
