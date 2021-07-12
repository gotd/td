module github.com/gotd/td/examples/gif-download

go 1.16

require (
	github.com/gotd/contrib v0.9.0
	github.com/gotd/td v0.43.0
	go.uber.org/atomic v1.8.0
	go.uber.org/zap v1.18.1
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace github.com/gotd/td => ./../..
