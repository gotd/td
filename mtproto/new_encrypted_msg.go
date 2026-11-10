package mtproto

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/proto"
)

func (c *Conn) newEncryptedMessage(id int64, seq int32, payload bin.Encoder, b *bin.Buffer) error {
	s := c.session()

	// TODO(tdakkota): Smarter gzip.
	// 	1) Generate Length() method for every encoder, to count length without encoding.
	// 	2) Re-use buffer instead of using yet one.
	// 	3) Do not send proto.GZIP if gzipped size is equal or bigger.
	var (
		d      crypto.EncryptedMessageData
		logger = c.log
	)
	if c.compressThreshold <= 0 && c.wire == nil {
		if obj, ok := payload.(interface{ TypeID() uint32 }); ok {
			logger = c.logWithTypeID(obj.TypeID())
		}
		d = crypto.EncryptedMessageData{
			SessionID: s.ID,
			Salt:      s.Salt,
			MessageID: id,
			SeqNo:     seq,
			Message:   payload,
		}
	} else {
		payloadBuf := bufPool.Get()
		defer bufPool.Put(payloadBuf)
		if err := payload.Encode(payloadBuf); err != nil {
			return errors.Wrap(err, "encode payload")
		}

		logger = c.logWithType(payloadBuf)
		body := payloadBuf
		compress := c.compressThreshold > 0 && payloadBuf.Len() > c.compressThreshold
		switch {
		case compress && c.wire == nil:
			d = crypto.EncryptedMessageData{
				SessionID: s.ID,
				Salt:      s.Salt,
				MessageID: id,
				SeqNo:     seq,
				Message:   proto.GZIP{Data: payloadBuf.Raw()},
			}
		default:
			if compress {
				// Tracing: encode gzip_packed separately to report the body
				// exactly as sent.
				gzipBuf := bufPool.Get()
				defer bufPool.Put(gzipBuf)
				if err := (proto.GZIP{Data: payloadBuf.Raw()}).Encode(gzipBuf); err != nil {
					return errors.Wrap(err, "encode gzip")
				}
				body = gzipBuf
			}
			if c.wire != nil {
				c.traceMessage(WireOut, s.Key, s.Salt, s.ID, id, seq, body.Raw())
			}
			d = crypto.EncryptedMessageData{
				SessionID:              s.ID,
				Salt:                   s.Salt,
				MessageID:              id,
				SeqNo:                  seq,
				MessageDataLen:         int32(body.Len()), // #nosec G115: bounded by the message size limit
				MessageDataWithPadding: body.Buf,
			}
		}
	}

	logger.Debug(context.Background(), "Request", log.Int64("msg_id", id))
	if err := c.cipher.Encrypt(s.Key, d, b); err != nil {
		return errors.Wrap(err, "encrypt")
	}

	return nil
}
