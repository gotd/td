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

// UberStyleErrors detects error messages like
//
// 	xerrors.Errorf("failed to do something: %w", err)
//
// but you should avoid "failed to" and use
//
// 	xerrors.Errorf("do something: %w", err)
//
// according to https://github.com/uber-go/guide/blob/master/style.md#error-wrapping.
func UberStyleErrors(m dsl.Matcher) {
	m.Match("$pkg.Errorf($msg, $*msg_args)").Where(
		m["msg"].Text.Matches(`"failed to.*"`),
	).Report("Avoid phrases like \"failed to\"")

	m.Match("errors.New($msg)").Where(
		m["msg"].Text.Matches(`"failed to.*"`),
	).Report("Avoid phrases like \"failed to\"")

	m.Match("xerrors.New($msg)").Where(
		m["msg"].Text.Matches(`"failed to.*"`),
	).Report("Avoid phrases like \"failed to\"")
}

// UnnecessaryErrorFormat detects unnecessary error formatting like
//
// 	xerrors.Errorf("error")
//
// and suggests instead
//
// 	xerrors.New("error")
//
func UnnecessaryErrorFormat(m dsl.Matcher) {
	m.Match("fmt.Errorf($msg)").Suggest("errors.New($msg)")
	m.Match("xerrors.Errorf($msg)").Suggest("xerrors.New($msg)")
}

// PreferXerrors suggests replace expressions like
//
// 	errors.As
//
// to
//
// 	xerrors.As
//
func PreferXerrors(m dsl.Matcher) {
	m.Match("errors.$method($*args)").Where(
		m["method"].Text.Matches("Is|As"),
	).Suggest("xerrors.$method($*args)")
}
