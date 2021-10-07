package mtproto

import (
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/internal/proto"
)

func (c *Conn) newEncryptedMessage(id int64, seq int32, payload bin.Encoder, b *bin.Buffer) error {
	s := c.session()

	// TODO(tdakkota): Smarter gzip.
	// 	1) Generate Length() method for every encoder, to count length without encoding.
	// 	2) Re-use buffer instead of using yet one.
	// 	3) Do not send proto.GZIP if gzipped size is equal or bigger.
	var (
		d   crypto.EncryptedMessageData
		log = c.log
	)
	if c.compressThreshold <= 0 {
		if obj, ok := payload.(interface{ TypeID() uint32 }); ok {
			log = c.logWithTypeID(obj.TypeID())
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
			return xerrors.Errorf("encode payload: %w", err)
		}

		log = c.logWithType(payloadBuf)
		if payloadBuf.Len() > c.compressThreshold {
			d = crypto.EncryptedMessageData{
				SessionID: s.ID,
				Salt:      s.Salt,
				MessageID: id,
				SeqNo:     seq,
				Message:   proto.GZIP{Data: payloadBuf.Raw()},
			}
		} else {
			d = crypto.EncryptedMessageData{
				SessionID:              s.ID,
				Salt:                   s.Salt,
				MessageID:              id,
				SeqNo:                  seq,
				MessageDataLen:         int32(payloadBuf.Len()),
				MessageDataWithPadding: payloadBuf.Buf,
			}
		}
	}

	log.Debug("Request", zap.Int64("msg_id", id))
	if err := c.cipher.Encrypt(s.Key, d, b); err != nil {
		return xerrors.Errorf("encrypt: %w", err)
	}

	return nil
}
