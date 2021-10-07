module github.com/nnqq/td/examples/bg-run

go 1.16

require (
	github.com/gotd/contrib v0.9.1-0.20210712180501-4e445979e6df
	github.com/nnqq/td v0.51.1
	go.uber.org/zap v1.19.1
)

replace github.com/nnqq/td => ./../..
