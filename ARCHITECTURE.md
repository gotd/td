# Architecture

General architecture and description of gotd.

Subprojects:

| Name                                     | Description                                                          |
| ---------------------------------------- | -------------------------------------------------------------------- |
| [getdoc](https://github.com/gotd/getdoc) | Documentation extraction (method descriptions, error messages, etc.) |
| [tl](https://github.com/gotd/tl)         | Schema parser and writer                                             |
| [ige](https://github.com/gotd/ige)       | AES-IGE block cipher for crypto                                      |

## Code generation

See the `internal/gen` package for code generation implementation.
We use `text/template`.

Generated packages:

| Name          | Description               |
| ------------- | ------------------------- |
| `tg`          | The latest Telegram layer |
| `tg/e2e`      | Secret chats schema       |
| `internal/mt` | MTProto schema            |

### Pipeline

1. Schema is parsed from `_schema/telegram.tl`
2. Embedded docs are loaded from `getdoc` if available
3. Bindings (interim representation) are generated
4. Type definitions are generated
5. Templates are executed with (4) in context
6. Source code is formatted and written to `tl_*_gen.go` files.

## Layers of abstraction

### Telegram

High-level API with helpers, like `downloader` or `uploader`.
Ideally, every telegram functionality should have sugared helper
that is convenient to use.

Can handle reconnects, DC migration, connection pooling, session management.

It is also possible to call methods directly from `tg.Client`, because
`telegram.Client` implements the `Invoker` interface:

```go
// Invoker can invoke raw MTProto RPC calls.
type Invoker interface {
	Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error
}
```

### MTProto

Low-level API abstracts out a single MTProto connection and handles it for a life cycle.
Implements background pings with keepalive.

Uses `internal/mt` (MTProto schema), `internal/proto` (MTProto-related implementation)
and `internal/rpc` (request-response handling, retries, acknowledgements) internally.

And it uses `internal/exchange` for the key exchange process and `internal/crypto` for encryption.

### Crypto

All cryptographical primitives that are used in key exchange or encryption are implemented in `internal/crypto` package.

Also, the `internal/crypto/srp` implements Secure Remote Password (2FA).

### Binary protocol

See `bin` package for implementation of MTProto basic types (de-)serialization.

We do a non-streaming approach, assuming that messages are fully available in memory,
so `bin.Buffer` is just a wrapper for byte slices that can read and write values.

We don't do a reflection-based approach, each serialization and deserialization is generated from the schema, and is constant.
