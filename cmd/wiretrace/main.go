// Binary wiretrace decodes and analyzes wire traces written by the traced
// tdesktop build (gotd/tdesktop fork, branch "trace") and by the
// github.com/gotd/td/wiretrace package, so gotd can copy the official
// client practices: which methods, in which order, how they are wrapped and
// when they are sent.
//
//	wiretrace decode   [--schema S] [--max-bytes N] [--dir in|out] [--method M] [--no-body] RUN
//	wiretrace timeline [--schema S] [--no-storage] [--kinds K,...] RUN
//	wiretrace stats    [--schema S] [--json] [--top N] RUN
//	wiretrace diff     [--schema S] [--schema-b S] [--timing] [--fields=false] A B
//
// RUN is a run directory with events-NNNN.jsonl and blobs-NNNN.bin, e.g.
// <tdesktop working dir>/trace/20260101T000000Z-1234 (UTC start time, pid).
//
// The schema is the api.tl and mtproto.tl of the traced build (tdesktop
// Telegram/SourceFiles/mtproto/scheme): --schema takes that directory or a
// comma-separated list of .tl files; without it, $WIRETRACE_SCHEMA and the
// run directory and its parents (and their "schema" subdirectories) are
// searched. decode and diff need it; timeline and stats only use it for
// names and fall back to the gotd tg/mt names.
//
// decode prints one JSON line per logical message: containers are split
// ("container" is the parent msg_id), wrappers are listed in "wrap"
// (["invokeWithLayer:233","initConnection"]), "method" is the innermost
// query, "req" and "caller" come from req.msg/req.caller, rpc_result gets
// "req_msg_id", the request "method", "result", "latency_ms" (wire) and
// "error", and "body" is the schema-decoded TL:
//
//	wiretrace decode --schema ~/src/tdesktop/Telegram/SourceFiles/mtproto/scheme RUN |
//		jq -c 'select(.dir=="out") | [.t, .method, .wrap]'
//
// timeline interleaves messages, requests, connection, key, session and
// storage events, with causes ("⇐ req#57 messages.getDialogs") and effects
// ("→ caused st.write map, req.new messages.getPeerDialogs"):
//
//	wiretrace timeline --no-storage RUN | less
//
// stats prints per-method counts and latency percentiles, rpc errors (with
// numbers masked), flood waits, reconnect reasons, per-DC sessions, messages
// and bytes, wire-level message kinds, callers and redactions.
//
// diff splits both runs into phases (connect, bind, init, auth, bootstrap,
// steady), maps each outgoing message to a shape token such as
// container[ping_delay_disconnect(75), invokeWithLayer(initConnection(help.getConfig))],
// aligns the tokens with LCS and reports messages only in A (-) or B (+),
// reordered methods, field differences of aligned messages (volatile fields
// such as msg_id, random_id, hashes and dates masked) and, with --timing,
// offsets, phase durations, ping cadence and reconnect backoff:
//
//	wiretrace diff --schema S --timing tdesktop-run gotd-run
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
)

const usage = `Usage: wiretrace <command> [flags] RUN...

Commands:
  decode     one JSON line per logical MTProto message
  timeline   human-readable network, storage and lifecycle events with causes
  stats      per-method counts, latencies, errors, flood waits, reconnects, bytes
  diff       compare the outgoing message shapes of two runs per phase

The schema (api.tl and mtproto.tl of the traced build) is taken from
--schema, $WIRETRACE_SCHEMA, or the run directory and its parents.
Run "wiretrace <command> -h" for command flags.
`

type command func(ctx context.Context, args []string, stdout, stderr io.Writer) error

var commands = map[string]command{
	"decode":   runDecode,
	"timeline": runTimeline,
	"stats":    runStats,
	"diff":     runDiff,
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		_, _ = fmt.Fprint(stderr, usage)
		return 2
	}
	cmd, ok := commands[args[0]]
	if !ok {
		if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
			_, _ = fmt.Fprint(stdout, usage)
			return 0
		}
		_, _ = fmt.Fprintf(stderr, "wiretrace: unknown command %q\n\n%s", args[0], usage)
		return 2
	}
	if err := cmd(ctx, args[1:], stdout, stderr); err != nil {
		if err == errUsage {
			return 2
		}
		_, _ = fmt.Fprintf(stderr, "wiretrace %s: %v\n", args[0], err)
		return 1
	}
	return 0
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	code := run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	cancel()
	os.Exit(code)
}
