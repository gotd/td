package proto

import (
	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
)

const ResultTypeID = 0xf35c6d01

type Result struct {
	RequestMessageID crypto.MessageID
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
		r.RequestMessageID = crypto.MessageID(v)
	}
	r.Result = append(r.Result[:0], b.Buf...)
	return nil
}
