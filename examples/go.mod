module github.com/gotd/td/examples

go 1.21

toolchain go1.21.0

require (
	github.com/cockroachdb/pebble v0.0.0-20220107203702-aa376a819bf6
	github.com/go-faster/errors v0.7.1
	github.com/gotd/contrib v0.19.0
	github.com/gotd/td v0.83.0
	github.com/joho/godotenv v1.5.1
	go.etcd.io/bbolt v1.3.7
	go.uber.org/atomic v1.11.0
	go.uber.org/zap v1.27.0
	golang.org/x/sync v0.7.0
	golang.org/x/term v0.19.0
	golang.org/x/time v0.3.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cockroachdb/errors v1.8.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20190617123548-eb05cc24525f // indirect
	github.com/cockroachdb/redact v1.0.8 // indirect
	github.com/cockroachdb/sentry-go v0.6.1-cockroachdb.2 // indirect
	github.com/go-faster/jx v1.1.0 // indirect
	github.com/go-faster/xor v1.0.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gotd/ige v0.2.2 // indirect
	github.com/gotd/neo v0.1.5 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.22.0 // indirect
	golang.org/x/exp v0.0.0-20230116083435-1de6713980de // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	nhooyr.io/websocket v1.8.11 // indirect
	rsc.io/qr v0.2.0 // indirect
)

replace github.com/gotd/td => ./..
