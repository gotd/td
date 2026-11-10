// Package tljson is a schema-driven generic TL decoder: it decodes binary TL
// into ordered JSON objects using api.tl and mtproto.tl parsed with
// github.com/gotd/tl, so it works for any layer without generated code.
//
// It handles flags (including flags2 and true bits), boxed and bare vectors,
// strings and bytes, int128/int256, double, generic queries
// (invokeWithLayer, initConnection, invokeAfterMsg, ...), rpc_result with
// the result type resolved by req_msg_id, msg_container, gzip_packed, the
// tracer redaction marker and unknown constructors (partial output).
//
// Output values: "_" is the constructor or function name, TL long is a
// decimal string, bytes are base64, int128/int256 are hex.
package tljson
