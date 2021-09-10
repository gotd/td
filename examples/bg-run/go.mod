module github.com/gotd/td/examples/bg-run

go 1.16

require (
	github.com/gotd/contrib v0.9.1-0.20210712180501-4e445979e6df
	github.com/gotd/td v0.50.0
	go.uber.org/zap v1.19.1
)

replace github.com/gotd/td => ./../..
