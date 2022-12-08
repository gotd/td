module github.com/gotd/td/examples/gif-download

go 1.16

require (
	github.com/go-faster/errors v0.6.1
	github.com/gotd/contrib v0.13.0
	github.com/gotd/td v0.60.0
	go.uber.org/atomic v1.10.0
	go.uber.org/zap v1.24.0
	golang.org/x/crypto v0.4.0
	golang.org/x/sync v0.1.0
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11
)

replace github.com/gotd/td => ./../..
