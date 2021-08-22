# Roadmap

## Q3

### Update key exchange

Updated key exchange protocol should be implemented

### Initial tracing

Add basic OpenTelemetry spans

### Bot API Types

All Bot API Types should be parsed
* Parser from docs
* OpenAPI v3 generated spec from parsed types
* Go structs generated from OpenAPI v3 spec

## Q4

### Sugared client

The tdlib-like client should be implemented to help users that are struggling with
raw API and a bunch of helpers.
* Caching for entities
  * Memory
  * External, e.g. in PostgreSQL or MongoDB
* Friendly wrappers for all objects
  * Dialogs
  * Users
  * Files
  * Photos

### Bot API

The `telegram-bot-api` compatible server should be implemented.

### Server

Embeddable Telegram server with limited functionality that can be used as a
server for `telegram-bot-api`.

### Bot API Client

Client for `telegram-bot-api`.

### Documentation
Create a static website for documentation for set of tools.

### Observability
* Advanced OpenTelemetry tracing
* Prometheus metrics

### Goals
* Implement ecosystem of tools for telegram in go
* Make it robust via extensive end-to-end testing
