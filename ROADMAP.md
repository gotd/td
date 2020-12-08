# Roadmap

Please use [project](https://github.com/gotd/td/projects) as source of truth for
features status.

## Gen

- [x] Vector
- [x] Fields
- [x] bytes type
- [ ] Non-bare vectors
- [x] Namespaces
- [x] Multiple output files
- [ ] Generated tests
- [ ] Generated examples
- [x] Boxes for class decode
- [x] RPC Requests
- [ ] Reduce signature of RPC requests with zero methods
- [ ] Vector as RPC Result
- [ ] RPC Error description from documentation error codes
- [ ] Automatically set optional (`#`) fields if they are not blank

## Client

- [x] Handle "bad server salt" error
- [x] Session storage
- [x] Automatic reconnect handling
- [ ] Background pings
- [ ] Replace zap with generic event listener, like [pebble.EventListener](https://pkg.go.dev/github.com/cockroachdb/pebble#EventListener)

## Testing
- [ ] tgtest, like http test (partially implemented)
- [ ] e2e tests (partially implemented)
