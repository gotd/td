# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

`gotd/td` is a pure-Go implementation of the Telegram [MTProto 2.0](https://core.telegram.org/mtproto)
protocol and a client built on top of it, for users and bots.

## Read first

- **[ARCHITECTURE.md](ARCHITECTURE.md)** — the layered design (binary → transport → crypto →
  `mtproto` connection → `pool` → high-level `telegram.Client`), the request lifecycle, the code
  generation pipeline, and what every package does. Read it before making non-trivial changes; it is
  the authoritative map of the codebase and is kept current.
- **[CONTRIBUTING.md](CONTRIBUTING.md)** — coding guidelines ([Uber style](https://github.com/uber-go/guide/blob/master/style.md)),
  testing-on-test-servers rules, and the allocation-testing expectations for hot paths.

## Reference implementations

`../../telegram` (i.e. `/src/telegram`) holds local clones of the official Telegram source — use them
to resolve protocol/API questions, error semantics, and undocumented behavior rather than guessing:

| Repo | Use for |
| --- | --- |
| `tdlib` | Canonical client behavior; the reference this project aims for parity with |
| `tdesktop` | Telegram Desktop; source of the `api.tl` schema and MTProto details |
| `android` | Official Android client (Java/C++) |
| `ios` | Official iOS client |
| `telegram-bot-api` | Bot API server; bot-specific semantics |
| `mtproxy` | MTProxy reference (faketls, obfuscation) |
| `tgcalls` | Voice/video call protocol reference |

### Filling

```bash
cd ../../
mkdir telegram
gh repo clone telegramdesktop/tdesktop
gh repo clone DrKLO/Telegram android
gh repo clone tdlib/td tdlib
gh repo clone tdlib/telegram-bot-api
gh repo clone TelegramMessenger/Telegram-iOS ios
gh repo clone TelegramMessenger/tgcalls
gh repo clone TelegramMessenger/MTProxy
```


## Commands

Go 1.25 (`go.mod`). CI runs against `oldstable` and `stable`.

```console
make test          # go test --timeout 5m -race ./...   (see go.test.sh)
make coverage      # coverage run, filters generated tl_*_gen.go (see go.coverage.sh)
make generate      # regenerate code from _schema/*.tl (go generate ./...)
make check_generated   # generate + `git diff --exit-code`; CI fails if generated code is stale
golangci-lint run  # lint; config in .golangci.yml
```

Run a single test / package:

```console
go test ./telegram/... -run TestClient_Connect
go test -race ./mtproto/...
```

## Things that bite

- **Generated code.** The `tg`, `tg/e2e`, `mt`, and `tgtrace` packages (and any `tl_*_gen.go` file)
  are generated from `_schema/*.tl` by `cmd/gotdgen` (templates in `gen/`). **Never edit `*_gen.go`
  by hand** — change the schema or the generator and run `make generate`. CI enforces this via
  `make check_generated`.

- **Schema updates** are deliberate: `make download_schema generate` pulls a new layer from
  tdesktop. Don't bump layers early (multiple in-flight layer versions cause breakage) — see
  CONTRIBUTING.md.

- **Offline testing.** Use `tgmock` (a mock `tg.Invoker`) to unit-test client helpers, and `tgtest`
  (an in-process pure-Go Telegram server, with `tgtest/cluster` for multi-DC) for end-to-end tests.
  Tests run with `-race`; never test against production Telegram servers. For time-dependent logic
  (pings, retries, timeouts) use the `clock` abstraction backed by `gotd/neo`, not `time` directly.

- **No runtime reflection.** `bin` serialization is non-streaming and reflection-free by design;
  keep hot paths allocation-light and verify with `testutil.ZeroAlloc` / `testutil.MaxAlloc` (see
  CONTRIBUTING.md "Testing allocations").

- **Context error semantics.** On context cancellation, return the wrapped `ctx.Err()`, never the
  underlying transport error.

- **`go-faster/errors.Wrap` does not short-circuit on nil.** `errors.Wrap(nil, msg)` returns a
  *non-nil* error (unlike `pkg/errors`). Guard with `if err != nil { return errors.Wrap(err, ...) }`
  before wrapping a result that may be nil.

- New external dependencies should be isolated in a subpackage so they don't leak into the core
  `clock`/`mtproto` tree (e.g. `clock/ntp` keeps `beevik/ntp` out of core).

## Commits & PRs

- [Conventional Commits](https://www.conventionalcommits.org/) (enforced by commitlint in CI), e.g.
  `feat(telegram): ...`, `fix(rpc): ...`.
- Branch off `main`; one focused PR per change.
