module github.com/gotd/td/examples

go 1.19

require (
	github.com/go-faster/errors v0.6.1
	github.com/gotd/contrib v0.15.0
	github.com/gotd/td v0.77.0
	go.uber.org/atomic v1.10.0
	go.uber.org/zap v1.24.0
	golang.org/x/crypto v0.8.0
	golang.org/x/sync v0.1.0
	golang.org/x/time v0.3.0
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
)

require (
	github.com/cenkalti/backoff/v4 v4.2.0 // indirect
	github.com/go-faster/jx v1.0.0 // indirect
	github.com/go-faster/xor v1.0.0 // indirect
	github.com/gotd/ige v0.2.2 // indirect
	github.com/gotd/neo v0.1.5 // indirect
	github.com/klauspost/compress v1.16.4 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	go.opentelemetry.io/otel v1.14.0 // indirect
	go.opentelemetry.io/otel/trace v1.14.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/term v0.7.0 // indirect
	nhooyr.io/websocket v1.8.7 // indirect
	rsc.io/qr v0.2.0 // indirect
)

replace github.com/gotd/td => ./..
