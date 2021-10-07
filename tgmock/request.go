package tgmock

import (
	"strconv"
	"strings"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/tg"
	"github.com/nnqq/td/tgerr"
)

func (i *Mock) request(fn func(b bin.Encoder)) *RequestBuilder {
	return &RequestBuilder{
		mock:     i,
		expectFn: fn,
		times:    1,
	}
}

// Expect creates builder of new expected call.
func (i *Mock) Expect() *RequestBuilder {
	return i.request(nil)
}

// ExpectFunc creates builder of new expected call with given input checker.
func (i *Mock) ExpectFunc(fn func(b bin.Encoder)) *RequestBuilder {
	return i.request(fn)
}

// ExpectCall creates builder of new expected call with given input.
func (i *Mock) ExpectCall(body bin.Encoder) *RequestBuilder {
	return i.request(func(got bin.Encoder) {
		i.assert.Equal(body, got)
	})
}

// RequestBuilder is builder of expected RPC request.
type RequestBuilder struct {
	mock     *Mock
	expectFn func(b bin.Encoder)
	times    int
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

// ThenTrue adds call tg.BoolTrue result to the end of call stack.
func (b *RequestBuilder) ThenTrue() *Mock {
	return b.result(&tg.BoolTrue{}, nil)
}

// ThenFalse adds call tg.BoolFalse result to the end of call stack.
func (b *RequestBuilder) ThenFalse() *Mock {
	return b.result(&tg.BoolFalse{}, nil)
}

// ThenErr adds call result to the end of call stack.
func (b *RequestBuilder) ThenErr(err error) *Mock {
	return b.result(nil, err)
}

// ThenRPCErr adds call result to the end of call stack.
func (b *RequestBuilder) ThenRPCErr(err *tgerr.Error) *Mock {
	return b.ThenErr(err)
}

// ThenMigrate adds call result to the end of call stack.
func (b *RequestBuilder) ThenMigrate(typ string, arg int) *Mock {
	t := strings.ToUpper(typ) + "_MIGRATE"
	return b.ThenRPCErr(&tgerr.Error{
		Code:     303,
		Message:  t + "_" + strconv.Itoa(arg),
		Type:     t,
		Argument: arg,
	})
}

// ThenFlood adds call result to the end of call stack.
func (b *RequestBuilder) ThenFlood(arg int) *Mock {
	return b.ThenRPCErr(&tgerr.Error{
		Code:     420,
		Message:  "FLOOD_WAIT_" + strconv.Itoa(arg),
		Type:     "FLOOD_WAIT",
		Argument: arg,
	})
}

// ThenUnregistered adds call result to the end of call stack.
func (b *RequestBuilder) ThenUnregistered() *Mock {
	return b.ThenRPCErr(&tgerr.Error{
		Code:    401,
		Message: "AUTH_KEY_UNREGISTERED",
		Type:    "AUTH_KEY_UNREGISTERED",
	})
}

func (b *RequestBuilder) result(r bin.Encoder, err error) *Mock {
	for i := 0; i < b.times; i++ {
		b.mock.add(HandlerFunc(func(id int64, body bin.Encoder) (bin.Encoder, error) {
			if b.expectFn != nil {
				b.expectFn(body)
			}
			return r, err
		}))
	}

	return b.mock
}
