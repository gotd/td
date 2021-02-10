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

## Binary protocol

See `bin` package for implementation of MTProto basic types (de-)serialization.

We use non-streaming approach, assuming that message is fully available in memory,
so `bin.Buffer` is just a wrapper for byte slice that can read and write values.

We do not use reflection-based approach, each serialization and de-serialization
is generated from schema and constant.
