package mtproto

import (
	"context"

	"github.com/go-faster/errors"
	"go.uber.org/zap"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/rpc"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

const maxBindTempAuthKeyAttempts = 3

func (c *Conn) bindTempAuthKey(ctx context.Context) error {
	var lastErr error
	for attempt := 1; attempt <= maxBindTempAuthKeyAttempts; attempt++ {
		err := c.bindTempAuthKeyAttempt(ctx)
		if err == nil {
			return nil
		}
		lastErr = err

		if tgerr.Is(err, tg.ErrEncryptedMessageInvalid) {
			if err := c.handleBindEncryptedMessageInvalid(attempt); err != nil {
				return err
			}
			continue
		}
		if tgerr.Is(err, "CONNECTION_NOT_INITED") {
			if err := c.handleBindConnectionNotInited(attempt); err != nil {
				return err
			}
			// Retry same bind call on transient ordering/race conditions.
			continue
		}
		return err
	}

	return errors.Wrap(lastErr, "temporary auth key bind retry limit reached")
}

func (c *Conn) bindTempAuthKeyAttempt(ctx context.Context) error {
	s := c.session()
	if s.Key.Zero() {
		return errors.New("temporary key is zero")
	}
	if s.PermKey.Zero() {
		return errors.New("permanent key is zero")
	}
	if s.ID == 0 {
		return errors.New("temporary session id is zero")
	}

	c.sessionMux.RLock()
	expiresAt := c.tempKeyExpiry
	c.sessionMux.RUnlock()
	if expiresAt == 0 {
		expiresAt = c.clock.Now().Unix() + int64(c.tempKeyTTL)
	}

	nonce, err := crypto.RandInt64(c.rand)
	if err != nil {
		return errors.Wrap(err, "generate nonce")
	}
	inner := &crypto.BindAuthKeyInner{
		Nonce:         nonce,
		TempAuthKeyID: s.Key.IntID(),
		PermAuthKeyID: s.PermKey.IntID(),
		TempSessionID: s.ID,
		ExpiresAt:     int(expiresAt),
	}

	// bindTempAuthKey is a content-related request and must consume normal
	// msg_id/seq_no progression to remain valid in current temp session.
	msgID, seqNo := c.nextMsgSeq(true)

	encryptedMessage, err := crypto.EncryptBindMessage(
		c.rand,
		s.PermKey,
		msgID,
		inner,
	)
	if err != nil {
		return errors.Wrap(err, "encrypt bind message")
	}

	req := &tg.AuthBindTempAuthKeyRequest{
		PermAuthKeyID:    s.PermKey.IntID(),
		Nonce:            nonce,
		ExpiresAt:        int(expiresAt),
		EncryptedMessage: encryptedMessage,
	}

	var result tg.BoolBox
	call := rpc.Request{
		MsgID:  msgID,
		SeqNo:  seqNo,
		Input:  req,
		Output: &result,
	}
	c.log.Debug("Binding temporary auth key",
		zap.Int64("temp_key_id", s.Key.IntID()),
		zap.Int64("perm_key_id", s.PermKey.IntID()),
		zap.Int64("temp_session_id", s.ID),
		zap.Int64("expires_at", expiresAt),
	)

	if err := c.rpc.Do(ctx, call); err != nil {
		var badMsgErr *badMessageError
		if errors.As(err, &badMsgErr) && badMsgErr.Code == codeIncorrectServerSalt {
			c.storeSalt(badMsgErr.NewSalt)
			c.salts.Reset()
			if err := c.rpc.Do(ctx, call); err != nil {
				return errors.Wrap(err, "invoke auth.bindTempAuthKey")
			}
		} else {
			return errors.Wrap(err, "invoke auth.bindTempAuthKey")
		}
	}

	if _, ok := result.Bool.(*tg.BoolTrue); !ok {
		return errors.New("temp auth key bind rejected")
	}
	c.log.Info("Temporary auth key bound",
		zap.Int64("temp_key_id", s.Key.IntID()),
		zap.Int64("perm_key_id", s.PermKey.IntID()),
		zap.Int64("expires_at", expiresAt),
	)
	return nil
}

func (c *Conn) handleBindEncryptedMessageInvalid(attempt int) error {
	// Quote (PFS): "if auth.bindTempAuthKey returns ENCRYPTED_MESSAGE_INVALID
	// and the permanent key was generated more than 60 seconds ago, both keys
	// should be dropped and generated again."
	// Link: https://core.telegram.org/api/pfs
	if c.permKeyOlderThan60s() {
		c.log.Warn("auth.bindTempAuthKey returned ENCRYPTED_MESSAGE_INVALID for old permanent key, dropping persisted PFS keys and reconnecting",
			zap.Int("attempt", attempt),
		)
		c.dropPFSKeys()
		return errors.Wrap(ErrPFSDropKeysRequired, "pfs keys dropped after ENCRYPTED_MESSAGE_INVALID")
	}

	if !c.permKeyAgeKnown() {
		// For restored sessions key age is unknown, so we first follow retry path.
		// If all retries fail, force key regeneration to avoid endless reconnect loop.
		if attempt >= maxBindTempAuthKeyAttempts {
			c.log.Warn("auth.bindTempAuthKey returned ENCRYPTED_MESSAGE_INVALID for key with unknown age after retries, dropping persisted PFS keys and reconnecting",
				zap.Int("attempt", attempt),
			)
			c.dropPFSKeys()
			return errors.Wrap(ErrPFSDropKeysRequired, "pfs keys dropped after repeated ENCRYPTED_MESSAGE_INVALID with unknown key age")
		}
		c.log.Warn("auth.bindTempAuthKey returned ENCRYPTED_MESSAGE_INVALID for key with unknown age, retrying bind before dropping keys",
			zap.Int("attempt", attempt),
		)
		return nil
	}

	// Quote (PFS): "Otherwise, the client should simply retry binding."
	// Link: https://core.telegram.org/api/pfs
	c.log.Warn("auth.bindTempAuthKey returned ENCRYPTED_MESSAGE_INVALID for fresh permanent key, retrying bind",
		zap.Int("attempt", attempt),
	)
	return nil
}

func (c *Conn) handleBindConnectionNotInited(attempt int) error {
	if attempt < maxBindTempAuthKeyAttempts {
		// In practice this can happen while server-side state is catching up;
		// avoid aggressive key reset and retry bind first.
		c.log.Warn("auth.bindTempAuthKey returned CONNECTION_NOT_INITED, retrying bind",
			zap.Int("attempt", attempt),
		)
		return nil
	}
	// After retries, reconnect whole transport/session and run full init path.
	c.log.Warn("auth.bindTempAuthKey returned CONNECTION_NOT_INITED after retries, reconnecting without dropping keys",
		zap.Int("attempt", attempt),
	)
	return errors.New("auth.bindTempAuthKey failed with CONNECTION_NOT_INITED after retries")
}

func (c *Conn) permKeyAgeKnown() bool {
	c.sessionMux.RLock()
	known := c.permKeyCreatedAt != 0
	c.sessionMux.RUnlock()
	return known
}

func (c *Conn) permKeyOlderThan60s() bool {
	c.sessionMux.RLock()
	createdAt := c.permKeyCreatedAt
	c.sessionMux.RUnlock()
	if createdAt == 0 {
		return false
	}
	return c.clock.Now().Unix()-createdAt > 60
}

func (c *Conn) dropPFSKeys() {
	c.sessionMux.Lock()
	c.authKey = crypto.AuthKey{}
	c.permKey = crypto.AuthKey{}
	c.permKeyCreatedAt = 0
	c.salt = 0
	c.sessionID = 0
	c.tempKeyExpiry = 0
	c.sessionMux.Unlock()
}
