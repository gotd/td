package telegram

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa" // #nosec
	"errors"
	"math/big"
	"sync/atomic"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
)

// createAuthKey generates new authorization key.
func (c *Client) createAuthKey(ctx context.Context) error {
	// 1. DH exchange initiation.
	nonce, err := crypto.RandInt128(c.rand)
	if err != nil {
		return xerrors.Errorf("client nonce generation: %w", err)
	}
	b := new(bin.Buffer)
	if err := c.newUnencryptedMessage(&mt.ReqPqMultiRequest{Nonce: nonce}, b); err != nil {
		return err
	}
	if err := c.conn.Send(ctx, b); err != nil {
		return xerrors.Errorf("write ReqPqMultiRequest: %w", err)
	}
	b.Reset()

	// 2. Server sends response of the form
	// resPQ#05162463 nonce:int128 server_nonce:int128 pq:string server_public_key_fingerprints:Vector long = ResPQ;
	if err := c.conn.Recv(ctx, b); err != nil {
		return xerrors.Errorf("read ResPQ message: %w", err)
	}
	plaintextMsg := proto.UnencryptedMessage{}
	if err := plaintextMsg.Decode(b); err != nil {
		return xerrors.Errorf("decode ResPQ message: %w", err)
	}

	b.ResetTo(plaintextMsg.MessageData)
	res := mt.ResPQ{}
	if err := res.Decode(b); err != nil {
		return err
	}
	if res.Nonce != nonce {
		return xerrors.New("ResPQ nonce mismatch")
	}

	// Selecting first public key that match fingerprint.
	var selectedPubKey *rsa.PublicKey
Loop:
	for _, fingerprint := range res.ServerPublicKeyFingerprints {
		for _, key := range c.rsaPublicKeys {
			if fingerprint == crypto.RSAFingerprint(key) {
				selectedPubKey = key
				break Loop
			}
		}
	}
	if selectedPubKey == nil {
		return xerrors.New("unable to select public key")
	}

	// The pq is a representation of a natural number (in binary big endian format).
	// SetBytes is also big endian.
	pq := big.NewInt(0).SetBytes(res.Pq)
	// Normally pq is less than or equal to 2^63-1.
	pqMax := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(63), nil)
	if pq.Cmp(pqMax) > 0 {
		return xerrors.New("server provided bad pq")
	}

	// 3. Client decomposes pq into prime factors such that p < q.
	// Performing proof of work.
	p, q, err := crypto.DecomposePQ(pq, c.rand)
	if err != nil {
		return xerrors.Errorf("decompose pq: %w", err)
	}

	// 4. Client sends query to server.
	// req_DH_params#d712e4be nonce:int128 server_nonce:int128 p:string q:string
	//   public_key_fingerprint:long encrypted_data:string = Server_DH_Params
	newNonce, err := crypto.RandInt256(c.rand)
	if err != nil {
		return xerrors.Errorf("generate new nonce: %w", err)
	}
	pqInnerData := &mt.PQInnerData{
		Pq:          pq.Bytes(),
		Nonce:       nonce,
		NewNonce:    newNonce,
		ServerNonce: res.ServerNonce,
		P:           p.Bytes(),
		Q:           q.Bytes(),
	}
	if err := pqInnerData.Encode(b); err != nil {
		return err
	}

	// `encrypted_data := RSA (data_with_hash, server_public_key);`
	encryptedData, err := crypto.RSAEncryptHashed(b.Buf, selectedPubKey, c.rand)
	if err != nil {
		return xerrors.Errorf("encrypted_data generation: %w", err)
	}
	reqDHParams := &mt.ReqDHParamsRequest{
		Nonce:                nonce,
		ServerNonce:          res.ServerNonce,
		P:                    p.Bytes(),
		Q:                    q.Bytes(),
		PublicKeyFingerprint: crypto.RSAFingerprint(selectedPubKey),
		EncryptedData:        encryptedData,
	}
	if err := c.newUnencryptedMessage(reqDHParams, b); err != nil {
		return err
	}
	if err := c.conn.Send(ctx, b); err != nil {
		return xerrors.Errorf("write ReqDHParamsRequest: %w", err)
	}
	b.Reset()

	// 5. Server responds with Server_DH_Params.
	if err := c.conn.Recv(ctx, b); err != nil {
		return xerrors.Errorf("read ServerDHParams message: %w", err)
	}
	if err := plaintextMsg.Decode(b); err != nil {
		return xerrors.Errorf("decode ServerDHParams message: %w", err)
	}

	b.ResetTo(plaintextMsg.MessageData)
	dhParams, err := mt.DecodeServerDHParams(b)
	if err != nil {
		return err
	}
	switch p := dhParams.(type) {
	case *mt.ServerDHParamsOk:
		// Success.
		if p.Nonce != nonce {
			return xerrors.New("ServerDHParamsOk nonce mismatch")
		}

		key, iv := crypto.TempAESKeys(newNonce.BigInt(), res.ServerNonce.BigInt())
		// Decrypting inner data.
		data, err := crypto.DecryptExchangeAnswer(p.EncryptedAnswer, key, iv)
		if err != nil {
			return xerrors.Errorf("exchange answer decrypt: %w", err)
		}
		b.ResetTo(data)

		innerData := mt.ServerDHInnerData{}
		if err := innerData.Decode(b); err != nil {
			return err
		}

		dhPrime := big.NewInt(0).SetBytes(innerData.DhPrime)
		g := big.NewInt(int64(innerData.G))
		gA := big.NewInt(0).SetBytes(innerData.GA)

		// 6. Random number b is computed:
		randMax := big.NewInt(0).SetBit(big.NewInt(0), 2048, 1)
		bParam, err := rand.Int(c.rand, randMax)
		if err != nil {
			return xerrors.Errorf("number b generation: %w", err)
		}
		// g_b = g^b mod dh_prime
		gB := big.NewInt(0).Exp(g, bParam, dhPrime)

		// Checking key exchange parameters.
		if err := crypto.CheckDHParams(dhPrime, g, gA, gB); err != nil {
			return xerrors.Errorf("key exchange failed: invalid params: %w", err)
		}

		clientInnerData := mt.ClientDHInnerData{
			ServerNonce: innerData.ServerNonce,
			Nonce:       innerData.Nonce,
			GB:          gB.Bytes(),
			// first attempt
			RetryID: 0,
		}
		b.Reset()
		if err := clientInnerData.Encode(b); err != nil {
			return err
		}
		clientEncrypted, err := crypto.EncryptExchangeAnswer(c.rand, b.Buf, key, iv)
		if err != nil {
			return xerrors.Errorf("exchange answer encrypt: %w", err)
		}

		setParamsReq := &mt.SetClientDHParamsRequest{
			Nonce:         nonce,
			ServerNonce:   reqDHParams.ServerNonce,
			EncryptedData: clientEncrypted,
		}
		b.Reset()
		if err := c.newUnencryptedMessage(setParamsReq, b); err != nil {
			return err
		}
		if err := c.conn.Send(ctx, b); err != nil {
			return xerrors.Errorf("write SetClientDHParamsRequest: %w", err)
		}

		// 7. Computing auth_key using formula (g_a)^b mod dh_prime
		authKey := big.NewInt(0).Exp(gA, bParam, dhPrime)

		b.Reset()
		if err := c.conn.Recv(ctx, b); err != nil {
			return xerrors.Errorf("read DhGenOk message: %w", err)
		}
		if err := plaintextMsg.Decode(b); err != nil {
			return xerrors.Errorf("decode DhGenOk message: %w", err)
		}
		b.ResetTo(plaintextMsg.MessageData)
		dhSetRes, err := mt.DecodeSetClientDHParamsAnswer(b)
		if err != nil {
			return xerrors.Errorf("decode DhGenOk answer: %w", err)
		}
		switch v := dhSetRes.(type) {
		case *mt.DhGenOk: // dh_gen_ok#3bcbf734
			var key crypto.AuthKey
			authKey.FillBytes(key[:])
			authKeyID := key.ID()

			// Checking received hash.
			var buf []byte
			buf = append(buf, newNonce[:]...)
			buf = append(buf, 1)
			buf = append(buf, sha(key[:])[0:8]...)
			nonceHash1 := sha(buf)[4:20]
			serverSalt := crypto.ServerSalt(newNonce, v.ServerNonce)

			if !bytes.Equal(nonceHash1, v.NewNonceHash1[:]) {
				return xerrors.New("key exchange verification failed: hash mismatch")
			}

			// Generating new session id and salt.
			sessionID, err := crypto.NewSessionID(c.rand)
			if err != nil {
				return err
			}

			c.authKey = crypto.AuthKeyWithID{AuthKey: key, AuthKeyID: authKeyID}
			atomic.StoreInt64(&c.session, sessionID)
			atomic.StoreInt64(&c.salt, serverSalt)

			return nil
		case *mt.DhGenRetry: // dh_gen_retry#46dc1fb9
			return xerrors.Errorf("retry required: %x", v.NewNonceHash2)
		case *mt.DhGenFail: // dh_gen_fail#a69dae02
			return xerrors.Errorf("dh_hen_fail: %x", v.NewNonceHash3)
		default:
			return xerrors.Errorf("unexpected result %T", v)
		}
	case *mt.ServerDHParamsFail:
		return errors.New("server respond with server_DH_params_fail")
	default:
		return xerrors.Errorf("unknown ")
	}
}
