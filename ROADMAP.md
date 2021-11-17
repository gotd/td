# Roadmap

## Q4 21
### Documentation

Generate a static websites for documentations.

### Observability

* Add basic OpenTelemetry spans _(from Q3)_
* Advanced OpenTelemetry tracing.
* Prometheus metrics.

### Updates engine
* ~~Refactor gap engine~~
* ~~Simplify initialization~~
* ~~Add example~~ and documentation

### ~~Update to go1.17~~
~~* Update go.mod to allow lazy load~~

## Q1 22

### Sugared client

Make the client simple for everyone with things like:

* Entity caching
  * In memory
  * External storage, for example in SQLite, PostgreSQL or MongoDB
* Friendly wrappers for all objects
  * Dialogs
  * Users
  * Files
  * Photos

### Bot API

An embeddable `telegram-bot-api` compatible server should be implemented.

### Server

An embeddable Telegram server with limited functionality that can be used as a
server for `telegram-bot-api`.

### Bot API Client

A client for `telegram-bot-api`.

### Goals

* Implement ecosystem of tools for Telegram in Go.
* Make it robust via extensive, extreme end-to-end testing, benchmarking and profiling.

## No Milestone

Features that have no milestone, but are likely to appear at some point.

* Extreme End-to-End
  * Advanced features
    * E2E encrypted chats
    * CDN support
    * Calls support
  * Components
    * Telegram Server in Go
    * Telegram Bot API in Go
    * Telegram Client in Go
    * Telegram Bot Client in Go
  * Tracing support in every component, on all layers of abstraction
    * MTProto level (withTraceID generic wrapper, like withoutUpdates)
    * Application level (spans in Server, Bot, Client, Events, Hooks)
    * Persistence level (spans in Database implementations of Client/Server persistence)
    * Background tasks level
    * Network level (eBPF)
* Continuous (daily, hourly consistent runs)
  * Profiling
    * Go
    * End-to-end
  * Benchmarks
    * Go
    * End-to-End
  * Fuzzing
* GitHub Actions (checks for PR's)
  * Benchmarks (Only significant changes)
  * API Backward Compatibility check
* Kubernetes integration
  * Telegram Bot API
  * Telegram Server
* Extreme performance
  * The [gnet](https://github.com/panjf2000/gnet) Event Loop for Server
  * Generated object pooling
  * Zero allocation AES-IGE implementation

## Changelog

### Q3 21

#### ~~Update key exchange~~

An updated key exchange protocol should be implemented. Done.

#### Initial tracing

Add basic OpenTelemetry spans (**moved to Q4**)

#### ~~Bot API Types~~

~~All Bot API Types should be parsed:~~

* ~~Parser from docs.~~
* ~~OpenAPI v3 generated spec from parsed types.~~
* ~~Go structs generated from OpenAPI v3 spec.~~
