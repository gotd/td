module github.com/gotd/td/examples/gif-download

go 1.16

require (
	github.com/go-faster/errors v0.6.0
	github.com/gotd/contrib v0.12.0
	github.com/gotd/td v0.54.1
	go.uber.org/atomic v1.9.0
	go.uber.org/zap v1.21.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
)

replace github.com/gotd/td => ./../..
