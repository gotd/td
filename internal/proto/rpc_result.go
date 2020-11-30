package proto

import (
	"github.com/gotd/td/bin"
)

const ResultTypeID = 0xf35c6d01

type Result struct {
	RequestMessageID int64
	Result           []byte
}

func (r *Result) Decode(b *bin.Buffer) error {
	if err := b.ConsumeID(ResultTypeID); err != nil {
		return err
	}
	{
		v, err := b.Long()
		if err != nil {
			return err
		}
		r.RequestMessageID = v
	}
	r.Result = append(r.Result[:0], b.Buf...)
	return nil
}
