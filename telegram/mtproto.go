package telegram

import (
	"context"

	"github.com/gotd/td/bin"
)

type MTProto interface {
	InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error
	Ping(ctx context.Context) error
	Close() error
}

var _ MTProto = mockInvoker(nil)

type mockInvoker func(input *bin.Buffer) (bin.Encoder, error)

func (m mockInvoker) Ping(ctx context.Context) error {
	return nil
}

func (m mockInvoker) InvokeRaw(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	// if _, ok := input.(proto.InvokeWithLayer); ok {
	// 	return nil
	// }

	buf := new(bin.Buffer)
	if err := input.Encode(buf); err != nil {
		return err
	}

	result, err := m(buf)
	if err != nil {
		return err
	}

	buf.Reset()
	if err := result.Encode(buf); err != nil {
		return err
	}

	return output.Decode(buf)
}

func (m mockInvoker) Close() error {
	return nil
}
