# Architecture

General architecture description of gotd project.

Subprojects:

|name|description|
|----|-----------|
|[getdoc](https://github.com/gotd/getdoc)| Documentation extraction (method description, error messages, etc) |
|[tl](https://github.com/gotd/tl)        | Schema parser and writer |
|[ige](https://github.com/gotd/ige)      | AES-IGE block cipher for crypto |

## Code generation

See `internal/gen` package for code generation implementation.
We use `text/template`.

Generated packages:

|name|description|
|----|-----------|
|`tg`          | Latest telegram layer |
|`tg/e2e`      | Secret chats schema |
|`internal/mt` | MTProto schema |

### Pipeline

1) Schema is parsed from `_schema/telegram.tl`
2) Embedded docs are loaded from `getdoc` if available
3) Bindings (interim representation) are generated
4) Type definitions are generated
5) Templates are executed with (4) in context
6) Source code is formatted and written to `tl_*_gen.go` files.


## Layers of abstraction

### Telegram

High level API with helpers, like `downloader` or `uploader`.
Ideally, every telegram functionality should have sugared helper
that is convenient to use.

Can handle reconnects, DC migration, connection pooling, session management.

Also, it is possible to call methods directly using `tg.Client`, because
`telegram.Client` implements `Invoker` interface:

```go
// Invoker can invoke raw MTProto rpc calls.
type Invoker interface {
	InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error
}
```

### MTProto

Low level API, abstracts out single MTProto connection and handles it life cycle.
Implements background pings with keepalive.

Uses `internal/mt` (MTProto schema),  `internal/proto` (MTProto-related implementation)
and `internal/rpc` (request-response handling, retries, acknowledgements) internally.

Also, use `internal/exchange` for key exchange process and `internal/crypto` for encryption.

### Crypto

All cryptographical primitives that are used in key exchange or encryption
are implemented in `internal/crypto` package.

Also, the `internal/crypto/srp` implements Secure Remote Password (2FA).

### Binary protocol

See `bin` package for implementation of MTProto basic types (de-)serialization.

We use non-streaming approach, assuming that message is fully available in memory,
so `bin.Buffer` is just a wrapper for byte slice that can read and write values.

We do not use reflection-based approach, each serialization and de-serialization
is generated from schema and is constant.
