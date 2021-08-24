# Roadmap

## Q3

### Update key exchange

An updated key exchange protocol should be implemented.

### Initial tracing

Add basic OpenTelemetry spans.

### Bot API Types

All Bot API Types should be parsed:

* Parser from docs.
* OpenAPI v3 generated spec from parsed types.
* Go structs generated from OpenAPI v3 spec.

## Q4

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

### Documentation

Generate a static websites for documentations.

### Observability

* Advanced OpenTelemetry tracing.
* Prometheus metrics.

### Goals

* Implement ecosystem of tools for Telegram in Go.
* Make it robust via extensive end-to-end testing.
