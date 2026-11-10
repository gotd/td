// Package wiretrace reads and writes wire traces: the plaintext of every
// encrypted MTProto message plus request, connection and storage lifecycle
// events, as written by the traced Telegram Desktop build (gotd/tdesktop
// fork, branch "trace").
//
// # Format
//
// A run directory holds segments:
//
//   - events-NNNN.jsonl: one JSON object per event. Every segment starts with
//     a "hdr" record (seq 0) with "v", "seg", "run", "app", "layer",
//     "api_sha256", "mtproto_sha256", "wall_utc", "mono_t", "pid" and
//     "flags". Other records have "seq", "t" (monotonic seconds), "th"
//     (thread) and "ev" (kind), plus kind-specific fields. 64-bit ids are
//     decimal strings, session ids are signed, constructor ids are
//     "0x%08x". Events emitted while a response or an update was handled
//     carry "cause":{"req","outer","type"}.
//   - blobs-NNNN.bin: "TDWT", u16 version 1, then records of u64 seq, u32 len,
//     u32 crc32c and the bytes. Events reference them with
//     "blob":{"seg","off","len","crc32c","sha256"}, where off points at the
//     bytes after the record header. Small bodies may also be inlined as
//     "b64".
//
// Segments rotate at 64 MiB of events or 256 MiB of blobs. When the total
// size cap is reached the oldest segments are deleted, except segment 0. If
// the writer queue overflows, a "dropped" record with "count" is written.
//
// "msg" events carry the body of one encrypted message (after decryption,
// before encryption) with the header fields (sdc, session, salt, msg_id,
// seqno, len) and a summary (type, method, wrap, inner, req_msg_id,
// result). Bodies of sensitive requests and responses are redacted in
// place, keeping lengths: [id][RedactedMarker][id][zeros...], listed in
// "redacted".
//
// Older tdesktop trace builds wrote numbers with the process locale decimal
// separator ("t":12,345678); ParseEvent repairs such lines.
//
// # Usage
//
//	run, err := wiretrace.Open(dir)
//	...
//	err = run.Events(func(ev *wiretrace.Event) error {
//		body, err := run.Blob(ev) // verified against crc32c and sha256
//		...
//	}, nil)
//
// Package wiretrace/tljson decodes TL bodies with the schema of the traced
// build, wiretrace/decode turns "msg" events into logical messages, and
// cmd/wiretrace is the command line tool (decode, timeline, stats, diff).
// Writer produces the same format, for tests and other emitters.
package wiretrace
