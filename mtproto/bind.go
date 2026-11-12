package mtproto

import (
	"context"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/log"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/rpc"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

const maxBindTempAuthKeyAttempts = 3

// bindRetryDelay is the delay before bind attempt n+1 after
// ENCRYPTED_MESSAGE_INVALID: a fresh permanent key may not have reached all
// servers yet, so retries need time (1 s, 2 s, 4 s).
func bindRetryDelay(attempt int) time.Duration {
	return time.Second << (attempt - 1)
}

func (c *Conn) bindTempAuthKey(ctx context.Context) error {
	var lastErr error
	for attempt := 1; attempt <= maxBindTempAuthKeyAttempts; attempt++ {
		err := c.bindTempAuthKeyAttempt(ctx)
		if err == nil {
			if c.sessionState != nil {
				c.sessionState.setLastBind(c.clock.Now().Unix())
			}
			return nil
		}
		lastErr = err

		if tgerr.Is(err, tg.ErrEncryptedMessageInvalid) {
			if err := c.handleBindEncryptedMessageInvalid(ctx, attempt); err != nil {
				return err
			}
			if attempt < maxBindTempAuthKeyAttempts {
				timer := c.clock.Timer(bindRetryDelay(attempt))
				select {
				case <-ctx.Done():
					clock.StopTimer(timer)
					return ctx.Err()
				case <-timer.C():
				}
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
	c.log.Debug(ctx, "Binding temporary auth key",
		log.Int64("temp_key_id", s.Key.IntID()),
		log.Int64("perm_key_id", s.PermKey.IntID()),
		log.Int64("temp_session_id", s.ID),
		log.Int64("expires_at", expiresAt),
	)

	var start time.Time
	if c.wire != nil {
		start = c.traceRequestStart(call)
	}
	err = c.rpc.Do(ctx, call)
	var badMsgErr *badMessageError
	if err != nil && errors.As(err, &badMsgErr) && badMsgErr.Code == codeIncorrectServerSalt {
		c.storeSalt(badMsgErr.NewSalt)
		c.salts.Reset()
		c.traceResend(call, "bad_server_salt")
		err = c.rpc.Do(ctx, call)
	}
	if c.wire != nil {
		c.traceRequestDone(call, start, err)
	}
	if err != nil {
		return errors.Wrap(err, "invoke auth.bindTempAuthKey")
	}

	if _, ok := result.Bool.(*tg.BoolTrue); !ok {
		return errors.New("temp auth key bind rejected")
	}
	c.log.Info(ctx, "Temporary auth key bound",
		log.Int64("temp_key_id", s.Key.IntID()),
		log.Int64("perm_key_id", s.PermKey.IntID()),
		log.Int64("expires_at", expiresAt),
	)
	return nil
}

// handleBindEncryptedMessageInvalid decides what to do when
// auth.bindTempAuthKey returns ENCRYPTED_MESSAGE_INVALID: retry the bind
// (nil), reconnect, or drop the permanent key (ErrPFSDropKeysRequired).
//
// The PFS docs say to drop both keys when the permanent key is older than
// 60 seconds, but that loses the authorization on a false positive. Like
// TDLib (Session::on_bind_result), the permanent key is immune when it is
// younger than 60 seconds (not yet propagated), or older than a day and was
// bound successfully within the last day. Otherwise it is checked with a
// request encrypted with it (checkPermKey) and dropped only if the server
// does not know it.
func (c *Conn) handleBindEncryptedMessageInvalid(ctx context.Context, attempt int) error {
	now := c.clock.Now().Unix()
	c.sessionMux.RLock()
	createdAt := c.permKeyCreatedAt
	c.sessionMux.RUnlock()
	var lastBind int64
	if c.sessionState != nil {
		lastBind = c.sessionState.lastBind()
	}
	age := now - createdAt
	immune := createdAt != 0 &&
		(age < 60 || (age > 86400 && lastBind > now-86400))
	if immune {
		if attempt >= maxBindTempAuthKeyAttempts {
			c.log.Warn(ctx, "auth.bindTempAuthKey returned ENCRYPTED_MESSAGE_INVALID after retries, reconnecting without dropping the permanent key",
				log.Int("attempt", attempt),
				log.Int64("perm_key_age", age),
			)
			return errors.Wrap(ErrPFSReconnectRequired, "temporary key bind failed with ENCRYPTED_MESSAGE_INVALID")
		}
		c.log.Warn(ctx, "auth.bindTempAuthKey returned ENCRYPTED_MESSAGE_INVALID, retrying bind",
			log.Int("attempt", attempt),
			log.Int64("perm_key_age", age),
		)
		return nil
	}

	check := c.permKeyCheck
	if check == nil {
		check = c.checkPermKey
	}
	unknown, err := check(ctx)
	if err != nil {
		c.log.Warn(ctx, "Permanent key check failed, reconnecting without dropping the key", log.Error(err))
		return errors.Wrap(ErrPFSReconnectRequired, "permanent key check failed")
	}
	if !unknown {
		if attempt >= maxBindTempAuthKeyAttempts {
			return errors.Wrap(ErrPFSReconnectRequired, "temporary key bind failed with ENCRYPTED_MESSAGE_INVALID, permanent key is valid")
		}
		c.log.Warn(ctx, "auth.bindTempAuthKey returned ENCRYPTED_MESSAGE_INVALID, permanent key is valid, retrying bind",
			log.Int("attempt", attempt),
		)
		return nil
	}
	c.log.Warn(ctx, "Server does not know the permanent key, dropping PFS keys",
		log.Int("attempt", attempt),
	)
	c.dropPFSKeys()
	return errors.Wrap(ErrPFSDropKeysRequired, "permanent key is unknown to the server")
}

// checkPermKey reports whether the server does not know the permanent key:
// it sends help.getNearestDc encrypted with the permanent key (without PFS)
// on a separate connection. The server answers -404 for an unknown key, which
// makes that connection generate a new key.
func (c *Conn) checkPermKey(ctx context.Context) (unknown bool, err error) {
	c.sessionMux.RLock()
	perm := c.permKey
	c.sessionMux.RUnlock()
	if perm.Zero() {
		return true, nil
	}
	if c.dialer == nil {
		return false, errors.New("no dialer")
	}
	probe := New(c.dialer, Options{
		DC:              c.dcID,
		PublicKeys:      c.rsaPublicKeys,
		Random:          c.rand,
		Logger:          c.log.Named("perm_key_check").Logger(),
		Clock:           c.clock,
		Key:             perm,
		ExchangeTimeout: c.exchangeTimeout,
		DialTimeout:     c.dialTimeout,
	})
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	runErr := probe.Run(ctx, func(ctx context.Context) error {
		var out tg.NearestDC
		err := probe.Invoke(ctx, &tg.HelpGetNearestDCRequest{}, &out)
		if _, ok := tgerr.As(err); ok {
			// Any RPC answer means the request was decrypted.
			return nil
		}
		return err
	})
	if probe.session().Key.ID != perm.ID {
		// -404: the probe had to generate a new key.
		return true, nil
	}
	if runErr != nil {
		return false, runErr
	}
	return false, nil
}

func (c *Conn) handleBindConnectionNotInited(attempt int) error {
	if attempt < maxBindTempAuthKeyAttempts {
		// In practice this can happen while server-side state is catching up;
		// avoid aggressive key reset and retry bind first.
		c.log.Warn(context.Background(), "auth.bindTempAuthKey returned CONNECTION_NOT_INITED, retrying bind",
			log.Int("attempt", attempt),
		)
		return nil
	}
	// After retries, reconnect whole transport/session and run full init path.
	c.log.Warn(context.Background(), "auth.bindTempAuthKey returned CONNECTION_NOT_INITED after retries, reconnecting without dropping keys",
		log.Int("attempt", attempt),
	)
	return errors.New("auth.bindTempAuthKey failed with CONNECTION_NOT_INITED after retries")
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
