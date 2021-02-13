package rpcmock

import (
	"strconv"
	"strings"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/mtproto"
)

func (i *Mock) request(body bin.Encoder) *RequestBuilder {
	return &RequestBuilder{
		mock:   i,
		expect: body,
		times:  1,
	}
}

// Expect creates builder of new expected call.
func (i *Mock) Expect() *RequestBuilder {
	return i.request(nil)
}

// ExpectCall creates builder of new expected call with given input.
func (i *Mock) ExpectCall(body bin.Encoder) *RequestBuilder {
	return i.request(body)
}

// RequestBuilder is builder of expected RPC request.
type RequestBuilder struct {
	mock   *Mock
	expect bin.Encoder
	times  int
}

// N sets count of expected calls.
func (b *RequestBuilder) N(times int) *RequestBuilder {
	b.times = times
	return b
}

// ThenResult adds call result to the end of call stack.
func (b *RequestBuilder) ThenResult(body bin.Encoder) *Mock {
	return b.result(body, nil)
}

// ThenErr adds call result to the end of call stack.
func (b *RequestBuilder) ThenErr(err error) *Mock {
	return b.result(nil, err)
}

// ThenRPCErr adds call result to the end of call stack.
func (b *RequestBuilder) ThenRPCErr(err *mtproto.Error) *Mock {
	return b.ThenErr(err)
}

// ThenMigrate adds call result to the end of call stack.
func (b *RequestBuilder) ThenMigrate(typ string, arg int) *Mock {
	t := strings.ToUpper(typ) + "_MIGRATE"
	return b.ThenRPCErr(&mtproto.Error{
		Code:     303,
		Message:  t + "_" + strconv.Itoa(arg),
		Type:     t,
		Argument: arg,
	})
}

// ThenFlood adds call result to the end of call stack.
func (b *RequestBuilder) ThenFlood(arg int) *Mock {
	return b.ThenRPCErr(&mtproto.Error{
		Code:     420,
		Message:  "FLOOD_WAIT_" + strconv.Itoa(arg),
		Type:     "FLOOD_WAIT",
		Argument: arg,
	})
}

// ThenUnregistered adds call result to the end of call stack.
func (b *RequestBuilder) ThenUnregistered() *Mock {
	return b.ThenRPCErr(&mtproto.Error{
		Code:    401,
		Message: "AUTH_KEY_UNREGISTERED",
		Type:    "AUTH_KEY_UNREGISTERED",
	})
}

func (b *RequestBuilder) result(r bin.Encoder, err error) *Mock {
	for i := 0; i < b.times; i++ {
		b.mock.add(HandlerFunc(func(id int64, body bin.Encoder) (bin.Encoder, error) {
			if b.expect != nil {
				b.mock.Assertions.Equal(b.expect, body)
			}
			return r, err
		}))
	}

	return b.mock
}
