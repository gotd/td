# Architecture

This document describes the general architecture of `gotd` — a pure Go
implementation of the [MTProto 2.0](https://core.telegram.org/mtproto)
protocol and the Telegram API client built on top of it.

The library is layered: each layer depends only on the layers below it and
exposes a small, well-defined interface to the layer above. From bottom to top
the layers are: binary serialization, transport, crypto, MTProto connection,
connection pool, and the high-level `telegram.Client`.

## Subprojects

`gotd` is split across several repositories under the
[gotd](https://github.com/gotd) and [go-faster](https://github.com/go-faster)
organizations.

| Name                                            | Description                                                          |
| ----------------------------------------------- | -------------------------------------------------------------------- |
| [getdoc](https://github.com/gotd/getdoc)        | Documentation extraction (method descriptions, error messages, etc.) |
| [tl](https://github.com/gotd/tl)                | TL schema parser and writer                                          |
| [ige](https://github.com/gotd/ige)              | AES-IGE block cipher used by the crypto layer                        |
| [neo](https://github.com/gotd/neo)              | Deterministic time source for tests (see `clock`)                    |
| [ogen](https://github.com/ogen-go/ogen)         | OpenAPI code generator (used by `tgacc`)                             |

## Layers of abstraction

```
                +---------------------------------------------------+
   high level   |  telegram.Client  (auth, updates, migration,      |
                |  uploads/downloads, message builders, peers, …)   |
                +---------------------------------------------------+
                          |  tg.Client (generated API) / Invoker
                +---------------------------------------------------+
   pool         |  pool.DC / pool.SetN  — connection pooling        |
                +---------------------------------------------------+
                          |  mtproto.Conn
                +---------------------------------------------------+
   connection   |  mtproto  — one MTProto connection life cycle,    |
                |  pings, salts, key exchange, (de)serialization    |
                +-----------+------------------+--------------------+
                            |                  |
                +-----------v----+   +---------v---------+   +---------------+
   support      |  rpc engine    |   |  crypto / exchange |   |  proto / bin  |
                |  (ack/retry)   |   |  (auth_key, AES)   |   |  primitives   |
                +----------------+   +-------------------+   +---------------+
                          |                  |                       |
                +---------v------------------v-----------------------v---------+
   transport    |  transport  — codecs (abridged, intermediate, full, …),     |
                |  obfuscation, TCP / WebSocket connections                    |
                +--------------------------------------------------------------+
```

### High level: `telegram`

[`telegram`](telegram) is the package most users interact with. The central
type is [`telegram.Client`](telegram/client.go), created via
`telegram.NewClient(appID, appHash, telegram.Options{})` and driven with
`Client.Run(ctx, f)`: the client connects, performs the initial handshake and
runs `f` while the connection is alive, tearing everything down when `f`
returns.

It owns all of the behavior that the lower layers deliberately leave out:

- **Authentication** ([`telegram/auth`](telegram/auth)) — user and bot login,
  2FA via SRP, QR-code login, code/password flows.
- **Update handling** ([`telegram/updates`](telegram/updates)) — ordered,
  gap-aware processing of the Telegram update sequence (`pts`/`qts`/`seq`),
  difference recovery, and channel state.
- **Datacenter management** ([`telegram/dcs`](telegram/dcs)) — DC discovery,
  resolvers, plain/WebSocket/MTProxy protocols, and migration/redirect
  handling.
- **Connection pooling** — multiple connections per DC and per-CDN DC pools
  built on the `pool` package (`pool.go`, `cdn_pool_manager.go`).
- **Session storage** ([`session`](session)) — persisting and restoring the
  `auth_key` and DC info so logins survive restarts.
- **Middlewares** (`middleware.go`) — `Invoker` decorators for rate limiting,
  `FLOOD_WAIT` retries, tracing, etc.
- **Convenience helpers** — [`uploader`](telegram/uploader),
  [`downloader`](telegram/downloader), [`message`](telegram/message) builders,
  [`query`](telegram/query) pagination, [`peers`](telegram/peers),
  [`thumbnail`](telegram/thumbnail), and [`takeout`](telegram/takeout).

Any raw MTProto method can be called directly: `telegram.Client` implements the
`Invoker` interface, so the generated `tg.Client` (`client.API()`) is layered
straight on top of it.

```go
// Invoker can invoke raw MTProto RPC calls.
type Invoker interface {
	Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error
}
```

### Pool: `pool`

[`pool`](pool) implements per-datacenter connection pools (`pool.DC`). It
acquires and releases connections, replaces dead ones, and load-balances
requests across them. `telegram.Client` uses it for both the primary DC and CDN
DCs. The pool is transport-agnostic: it is parameterized by a connection
constructor and only deals in the `Conn` `Invoke` interface.

### Connection: `mtproto`

[`mtproto`](mtproto) abstracts a **single** MTProto connection and manages its
full life cycle (`conn.go`):

- Runs the [key exchange](exchange) on first connect to obtain the `auth_key`,
  or reuses a stored one.
- Maintains the server salt (`salt.go`, `salts/`), session, and message-id
  generation (`message_id.go`).
- Sends background pings with keepalive (`ping.go`) and reconnects on failure.
- (De)serializes, encrypts, and routes incoming messages to handlers
  (`handle_*.go`, `read.go`, `write.go`), including containers, gzip, RPC
  results, bad-msg notifications, future salts, and session-created events.
- Supports Perfect Forward Secrecy via temporary keys (`pfs.go`, `bind.go`).

It is deliberately **single-connection**: reconnection across DCs, pooling, and
update handling live in higher layers.

### RPC engine: `rpc`

[`rpc`](rpc) is the request/response engine (`engine.go`). It matches RPC
results to in-flight requests by message id, handles acknowledgements
(`ack.go`), retransmits unacknowledged requests, and supports context-based
cancellation. `mtproto.Conn` delegates request bookkeeping to it.

### Crypto: `crypto` and `exchange`

[`crypto`](crypto) implements every cryptographic primitive used by MTProto:
the `auth_key` derivation, AES-IGE message encryption/decryption (`cipher*.go`),
key/message-key derivation (`kdf_v1.go`, `key.go`), RSA with the custom padding
used during exchange (`rsa*.go`), PQ factorization (`pq.go`), Diffie-Hellman
checks (`check_dh.go`, `dh.go`), secure PRNG (`rand*.go`), and the vendored
public keys (`public_keys.go`).

[`crypto/srp`](crypto/srp) implements Secure Remote Password for 2FA.

[`exchange`](exchange) drives the [auth key generation
protocol](https://core.telegram.org/mtproto/auth_key) itself, with both client
(`client_flow.go`) and server (`server_flow.go`) flows — the server flow is used
by the in-process test server.

### Binary protocol: `bin` and `proto`

[`bin`](bin) implements (de)serialization of the basic TL/MTProto wire types.
It uses a **non-streaming** approach: a message is assumed to be fully in
memory, so `bin.Buffer` is a thin wrapper over a byte slice that reads and
writes values. There is **no runtime reflection** — every type's
`Encode`/`Decode` is generated from the schema and is therefore constant-cost.

[`proto`](proto) builds the MTProto 2.0 message primitives on top of `bin`:
message containers (`container.go`), gzip packing (`gzip.go`), message ids
(`message_id.go`), RPC result wrapping (`rpc_result.go`), and unencrypted
messages used during key exchange (`unencrypted_message.go`).

### Transport: `transport`

[`transport`](transport) contains the MTProto transport implementations and
[codecs](https://core.telegram.org/mtproto/mtproto-transports): abridged,
intermediate, padded-intermediate, and full (`codec.go`, `protocol.go`). It
provides both TCP and [WebSocket](transport/websocket.go) connections (the
latter makes the client usable from WASM) and the obfuscation wrapper
(`obfuscated.go`). [`mtproxy`](mtproxy) builds MTProxy support
(`faketls`, `obfuscated2`) on the same abstractions.

## Generated code: `tg`, `tg/e2e`, `mt`

A large fraction of the codebase is generated from [TL
schemas](https://core.telegram.org/mtproto/TL) in [`_schema`](_schema). The
generated `tg` package alone is hundreds of thousands of lines of constant,
reflection-free (de)serialization and typed method wrappers.

Generated packages:

| Package          | Schema                  | Description                          |
| ---------------- | ----------------------- | ------------------------------------ |
| [`tg`](tg)       | `_schema/telegram.tl`   | The latest Telegram API layer        |
| [`tg/e2e`](tg/e2e) | `_schema/encrypted.tl`| Secret-chats (end-to-end) schema     |
| [`mt`](mt)       | `_schema/mt.tl`         | MTProto service-message schema       |
| [`tgtrace`](tgtrace) | `_schema/trace.tl`  | Tracing schema                       |

### Generation pipeline

Code generation lives in [`gen`](gen) and is driven by the
[`cmd/gotdgen`](cmd/gotdgen) command (see the `//go:generate` directives in
[`td.go`](td.go)). It uses `text/template` (`gen/_template`, `templates.go`):

1. The schema is parsed from `_schema/*.tl` using the
   [`gotd/tl`](https://github.com/gotd/tl) parser.
2. Embedded docs are loaded from [`getdoc`](https://github.com/gotd/getdoc) when
   available (method descriptions, error messages).
3. *Bindings* — an interim representation — are generated (`make_bindings.go`).
4. Type definitions (structures, interfaces, vectors) are built
   (`make_structures.go`, `make_interfaces.go`, `make_vector.go`,
   `make_field.go`).
5. Templates are executed against (4) to produce Go source.
6. The result is gofmt-formatted and written to `tl_*_gen.go` files
   (`write_source.go`).

To regenerate, run `go generate ./...` (or `make generate`).

## Supporting packages

| Package                  | Description                                                        |
| ------------------------ | ----------------------------------------------------------------- |
| [`session`](session)     | Pluggable session storage (file, memory, JS); imports from TDesktop/Telethon |
| [`tdp`](tdp)             | Pretty-printing/formatting of generated MTProto types             |
| [`tdsync`](tdsync)       | Concurrency helpers (supervisor, ready/reset, backoff)            |
| [`tmap`](tmap)           | Type-id → constructor maps for decoding                           |
| [`tgerr`](tgerr)         | Telegram RPC error parsing and matching (e.g. `FLOOD_WAIT`)       |
| [`clock`](clock)         | Abstract time source (real or deterministic for tests)            |
| [`syncio`](syncio)       | Synchronized `io` wrappers                                        |
| [`fileid`](fileid)       | Bot-API style file-id encode/decode                               |
| [`constant`](constant)   | Telegram-defined constants                                        |
| [`oteltg`](oteltg)       | OpenTelemetry instrumentation                                     |
| [`wsutil`](wsutil)       | WebSocket utilities                                               |
| [`pool`](pool)           | Generic connection-pool primitives                                |

## Testing infrastructure

The project is tested at every layer (see the [README](README.md) for the full
list). The notable pieces:

- [`tgtest`](tgtest) — an in-process Telegram server written in pure Go,
  enabling end-to-end tests without the real network. `tgtest/cluster` spins up
  a multi-DC setup; `tgtest/services` provides server-side behavior.
- [`tgmock`](tgmock) — a mock `tg.Invoker` for unit-testing code that issues
  RPC calls.
- [`testutil`](testutil) and [`clock`](clock) (backed by
  [`gotd/neo`](https://github.com/gotd/neo)) — deterministic time for testing
  timeouts, pings, and retries.
- [`_fuzz`](_fuzz) — fuzzing corpora for message handling, the key-exchange
  flow, and RSA.
- End-to-end tests against the real Telegram server run in CI, and a 24/7 canary
  bot exercises reconnects, update handling, memory, and performance in
  production.

## Request lifecycle

Putting the layers together, a typical `client.API().SomeMethod(ctx, …)` call
flows as follows:

1. The generated `tg.Client` method serializes the request with `bin` and calls
   `Invoke` on the `telegram.Client`.
2. Middlewares run (rate limiting, flood-wait retry, tracing).
3. `telegram.Client` selects/acquires a connection from the `pool` for the
   target DC (migrating or redirecting to a CDN DC if required).
4. `mtproto.Conn` wraps the payload in an MTProto message (`proto`), encrypts it
   with the session `auth_key` (`crypto`), and writes it over the `transport`
   codec.
5. The `rpc` engine tracks the message id, sends acks, and retransmits if
   needed.
6. The response is read, decrypted, decoded, and routed back to the waiting
   caller; updates are dispatched to the `telegram/updates` manager and on to
   the user's `UpdateHandler`.
