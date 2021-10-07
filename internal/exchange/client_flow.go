package exchange

import (
	"context"
	"crypto/rand"
	"math/big"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/internal/mt"
	"github.com/nnqq/td/internal/proto"
)

// Run runs client-side flow.
func (c ClientExchange) Run(ctx context.Context) (ClientExchangeResult, error) {
	// 1. DH exchange initiation.
	nonce, err := crypto.RandInt128(c.rand)
	if err != nil {
		return ClientExchangeResult{}, xerrors.Errorf("client nonce generation: %w", err)
	}
	b := new(bin.Buffer)

	c.log.Debug("Sending ReqPqMultiRequest")
	if err := c.writeUnencrypted(ctx, b, &mt.ReqPqMultiRequest{Nonce: nonce}); err != nil {
		return ClientExchangeResult{}, xerrors.Errorf("write ReqPqMultiRequest: %w", err)
	}

	// 2. Server sends response of the form
	// resPQ#05162463 nonce:int128 server_nonce:int128 pq:string server_public_key_fingerprints:Vector long = ResPQ;
	var res mt.ResPQ
	if err := c.readUnencrypted(ctx, b, &res); err != nil {
		return ClientExchangeResult{}, xerrors.Errorf("read ResPQ response: %w", err)
	}
	c.log.Debug("Received server ResPQ")
	if res.Nonce != nonce {
		return ClientExchangeResult{}, xerrors.New("ResPQ nonce mismatch")
	}
	serverNonce := res.ServerNonce

	// Selecting first public key that match fingerprint.
	var selectedPubKey PublicKey
Loop:
	for _, key := range c.keys {
		f := key.Fingerprint()

		for _, fingerprint := range res.ServerPublicKeyFingerprints {
			if fingerprint == f {
				selectedPubKey = key
				break Loop
			}
		}
	}
	if selectedPubKey.Zero() {
		return ClientExchangeResult{}, ErrKeyFingerprintNotFound
	}

	// The pq is a representation of a natural number (in binary big endian format).
	// SetBytes is also big endian.
	pq := big.NewInt(0).SetBytes(res.Pq)
	// Normally pq is less than or equal to 2^63-1.
	pqMax := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(63), nil)
	if pq.Cmp(pqMax) > 0 {
		return ClientExchangeResult{}, xerrors.New("server provided bad pq")
	}

	start := c.clock.Now()
	// 3. Client decomposes pq into prime factors such that p < q.
	// Performing proof of work.
	p, q, err := crypto.DecomposePQ(pq, c.rand)
	if err != nil {
		return ClientExchangeResult{}, xerrors.Errorf("decompose pq: %w", err)
	}
	c.log.Debug("PQ decomposing complete", zap.Duration("took", c.clock.Now().Sub(start)))
	// Make a copy of p and q values to reduce allocations.
	pBytes := p.Bytes()
	qBytes := q.Bytes()

	// 4. Client sends query to server.
	// req_DH_params#d712e4be nonce:int128 server_nonce:int128 p:string q:string
	//   public_key_fingerprint:long encrypted_data:string = Server_DH_Params
	newNonce, err := crypto.RandInt256(c.rand)
	if err != nil {
		return ClientExchangeResult{}, xerrors.Errorf("generate new nonce: %w", err)
	}

	var encryptedData []byte
	if !selectedPubKey.UseRSAPad {
		pqInnerData := &mt.PQInnerData{
			Pq:          res.Pq,
			Nonce:       nonce,
			NewNonce:    newNonce,
			ServerNonce: serverNonce,
			P:           pBytes,
			Q:           qBytes,
		}
		b.Reset()
		if err := pqInnerData.Encode(b); err != nil {
			return ClientExchangeResult{}, err
		}

		// `encrypted_data := RSA (data_with_hash, server_public_key);`
		data, err := crypto.RSAEncryptHashed(b.Buf, selectedPubKey.RSA, c.rand)
		if err != nil {
			return ClientExchangeResult{}, xerrors.Errorf("encrypted_data generation: %w", err)
		}

		encryptedData = data
	} else {
		pqInnerData := &mt.PQInnerDataDC{
			Pq:          res.Pq,
			Nonce:       nonce,
			NewNonce:    newNonce,
			ServerNonce: serverNonce,
			P:           pBytes,
			Q:           qBytes,
			DC:          c.dc,
		}
		b.Reset()
		if err := pqInnerData.Encode(b); err != nil {
			return ClientExchangeResult{}, err
		}

		// `encrypted_data := RSA_PAD(data, server_public_key);`
		data, err := crypto.RSAPad(b.Buf, selectedPubKey.RSA, c.rand)
		if err != nil {
			return ClientExchangeResult{}, xerrors.Errorf("encrypted_data generation: %w", err)
		}

		encryptedData = data
	}

	reqDHParams := &mt.ReqDHParamsRequest{
		Nonce:                nonce,
		ServerNonce:          serverNonce,
		P:                    pBytes,
		Q:                    qBytes,
		PublicKeyFingerprint: selectedPubKey.Fingerprint(),
		EncryptedData:        encryptedData,
	}
	c.log.Debug("Sending ReqDHParamsRequest")
	if err := c.writeUnencrypted(ctx, b, reqDHParams); err != nil {
		return ClientExchangeResult{}, xerrors.Errorf("write ReqDHParamsRequest: %w", err)
	}

	// 5. Server responds with Server_DH_Params.
	if err := c.conn.Recv(ctx, b); err != nil {
		return ClientExchangeResult{}, xerrors.Errorf("read ServerDHParams message: %w", err)
	}
	c.log.Debug("Received server ServerDHParams")

	var plaintextMsg proto.UnencryptedMessage
	if err := plaintextMsg.Decode(b); err != nil {
		return ClientExchangeResult{}, xerrors.Errorf("decode ServerDHParams message: %w", err)
	}

	b.ResetTo(plaintextMsg.MessageData)
	dhParams, err := mt.DecodeServerDHParams(b)
	if err != nil {
		return ClientExchangeResult{}, err
	}
	switch p := dhParams.(type) {
	case *mt.ServerDHParamsOk:
		// Success.
		if p.Nonce != nonce {
			return ClientExchangeResult{}, xerrors.New("ServerDHParamsOk nonce mismatch")
		}
		if p.ServerNonce != serverNonce {
			return ClientExchangeResult{}, xerrors.New("ServerDHParamsOk server nonce mismatch")
		}

		key, iv := crypto.TempAESKeys(newNonce.BigInt(), serverNonce.BigInt())
		// Decrypting inner data.
		data, err := crypto.DecryptExchangeAnswer(p.EncryptedAnswer, key, iv)
		if err != nil {
			return ClientExchangeResult{}, xerrors.Errorf("exchange answer decrypt: %w", err)
		}
		b.ResetTo(data)

		innerData := mt.ServerDHInnerData{}
		if err := innerData.Decode(b); err != nil {
			return ClientExchangeResult{}, err
		}
		if innerData.Nonce != nonce {
			return ClientExchangeResult{}, xerrors.New("ServerDHInnerData nonce mismatch")
		}
		if innerData.ServerNonce != serverNonce {
			return ClientExchangeResult{}, xerrors.New("ServerDHInnerData server nonce mismatch")
		}

		dhPrime := big.NewInt(0).SetBytes(innerData.DhPrime)
		g := big.NewInt(int64(innerData.G))
		if err := crypto.CheckDH(innerData.G, dhPrime); err != nil {
			return ClientExchangeResult{}, xerrors.Errorf("check DH params: %w", err)
		}
		gA := big.NewInt(0).SetBytes(innerData.GA)

		// 6. Random number b is computed:
		randMax := big.NewInt(0).SetBit(big.NewInt(0), crypto.RSAKeyBits, 1)
		bParam, err := rand.Int(c.rand, randMax)
		if err != nil {
			return ClientExchangeResult{}, xerrors.Errorf("number b generation: %w", err)
		}
		// g_b = g^b mod dh_prime
		gB := big.NewInt(0).Exp(g, bParam, dhPrime)

		// Checking key exchange parameters.
		if err := crypto.CheckDHParams(dhPrime, g, gA, gB); err != nil {
			return ClientExchangeResult{}, xerrors.Errorf("key exchange failed: invalid params: %w", err)
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
			return ClientExchangeResult{}, err
		}
		clientEncrypted, err := crypto.EncryptExchangeAnswer(c.rand, b.Buf, key, iv)
		if err != nil {
			return ClientExchangeResult{}, xerrors.Errorf("exchange answer encrypt: %w", err)
		}

		setParamsReq := &mt.SetClientDHParamsRequest{
			Nonce:         nonce,
			ServerNonce:   reqDHParams.ServerNonce,
			EncryptedData: clientEncrypted,
		}
		c.log.Debug("Sending SetClientDHParamsRequest")
		if err := c.writeUnencrypted(ctx, b, setParamsReq); err != nil {
			return ClientExchangeResult{}, xerrors.Errorf("write SetClientDHParamsRequest: %w", err)
		}

		// 7. Computing auth_key using formula (g_a)^b mod dh_prime
		authKey := big.NewInt(0).Exp(gA, bParam, dhPrime)

		b.Reset()
		if err := c.conn.Recv(ctx, b); err != nil {
			return ClientExchangeResult{}, xerrors.Errorf("read DhGen message: %w", err)
		}
		c.log.Debug("Received server DhGen")

		if err := plaintextMsg.Decode(b); err != nil {
			return ClientExchangeResult{}, xerrors.Errorf("decode DhGen message: %w", err)
		}
		b.ResetTo(plaintextMsg.MessageData)
		dhSetRes, err := mt.DecodeSetClientDHParamsAnswer(b)
		if err != nil {
			return ClientExchangeResult{}, xerrors.Errorf("decode DhGen answer: %w", err)
		}
		switch v := dhSetRes.(type) {
		case *mt.DhGenOk: // dh_gen_ok#3bcbf734
			if v.Nonce != nonce {
				return ClientExchangeResult{}, xerrors.New("DhGenOk nonce mismatch")
			}
			if v.ServerNonce != serverNonce {
				return ClientExchangeResult{}, xerrors.New("DhGenOk server nonce mismatch")
			}

			var key crypto.Key
			authKey.FillBytes(key[:])
			authKeyID := key.ID()

			// Checking received hash.
			nonceHash1 := crypto.NonceHash1(newNonce, key)
			serverSalt := crypto.ServerSalt(newNonce, v.ServerNonce)

			if nonceHash1 != v.NewNonceHash1 {
				return ClientExchangeResult{}, xerrors.New("key exchange verification failed: hash mismatch")
			}

			// Generating new session id and salt.
			sessionID, err := crypto.NewSessionID(c.rand)
			if err != nil {
				return ClientExchangeResult{}, err
			}

			return ClientExchangeResult{
				AuthKey:    crypto.AuthKey{Value: key, ID: authKeyID},
				SessionID:  sessionID,
				ServerSalt: serverSalt,
			}, nil
		case *mt.DhGenRetry: // dh_gen_retry#46dc1fb9
			return ClientExchangeResult{}, xerrors.Errorf("retry required: %x", v.NewNonceHash2)
		case *mt.DhGenFail: // dh_gen_fail#a69dae02
			return ClientExchangeResult{}, xerrors.Errorf("dh_hen_fail: %x", v.NewNonceHash3)
		default:
			return ClientExchangeResult{}, xerrors.Errorf("unexpected SetClientDHParamsRequest result %T", v)
		}
	case *mt.ServerDHParamsFail:
		return ClientExchangeResult{}, xerrors.New("server respond with server_DH_params_fail")
	default:
		return ClientExchangeResult{}, xerrors.Errorf("unexpected ReqDHParamsRequest result %T", p)
	}
}
