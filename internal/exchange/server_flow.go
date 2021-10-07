package exchange

import (
	"context"
	"math/big"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/nnqq/td/bin"
	"github.com/nnqq/td/internal/crypto"
	"github.com/nnqq/td/internal/mt"
	"github.com/nnqq/td/internal/proto/codec"
)

// ServerExchangeError is returned when exchange fails due to
// some security or validation checks.
type ServerExchangeError struct {
	Code int32
	Err  error
}

// Error implements error.
func (s *ServerExchangeError) Error() string {
	return s.Err.Error()
}

// Unwrap implements error wrapper interface.
func (s *ServerExchangeError) Unwrap() error {
	return s.Err
}

func serverError(code int32, err error) error {
	return &ServerExchangeError{
		Code: code,
		Err:  err,
	}
}

// Run runs server-side flow.
// If b parameter is not nil, it will be used as first read message.
// Otherwise, it will be read from connection.
func (s ServerExchange) Run(ctx context.Context) (ServerExchangeResult, error) {
	wrapKeyNotFound := func(err error) error {
		return serverError(codec.CodeAuthKeyNotFound, err)
	}

	// 1. Client sends query to server
	//
	// req_pq_multi#be7e8ef1 nonce:int128 = ResPQ;
	var pqReq mt.ReqPqMultiRequest
	b := new(bin.Buffer)
	if err := s.readUnencrypted(ctx, b, &pqReq); err != nil {
		return ServerExchangeResult{}, err
	}
	s.log.Debug("Received client ReqPqMultiRequest")

	serverNonce, err := crypto.RandInt128(s.rand)
	if err != nil {
		return ServerExchangeResult{}, xerrors.Errorf("generate server nonce: %w", err)
	}

	// 2. Server sends response of the form
	//
	// resPQ#05162463 nonce:int128 server_nonce:int128 pq:string server_public_key_fingerprints:Vector long = ResPQ;
	pq, err := s.rng.PQ()
	if err != nil {
		return ServerExchangeResult{}, xerrors.Errorf("generate pq: %w", err)
	}

	s.log.Debug("Sending ResPQ", zap.String("pq", pq.String()))
	if err := s.writeUnencrypted(ctx, b, &mt.ResPQ{
		Pq:          pq.Bytes(),
		Nonce:       pqReq.Nonce,
		ServerNonce: serverNonce,
		ServerPublicKeyFingerprints: []int64{
			s.key.Fingerprint(),
		},
	}); err != nil {
		return ServerExchangeResult{}, err
	}

	// 4. Client sends query to server
	//
	// req_DH_params#d712e4be nonce:int128 server_nonce:int128 p:string
	//  q:string public_key_fingerprint:long encrypted_data:string = Server_DH_Params
	var dhParams mt.ReqDHParamsRequest
	if err := s.readUnencrypted(ctx, b, &dhParams); err != nil {
		return ServerExchangeResult{}, err
	}
	s.log.Debug("Received client ReqDHParamsRequest")

	var innerData mt.PQInnerData
	{
		if !s.key.UseRSAPad {
			r, err := crypto.RSADecryptHashed(dhParams.EncryptedData, s.key.RSA)
			if err != nil {
				return ServerExchangeResult{}, wrapKeyNotFound(err)
			}
			b.ResetTo(r)
		} else {
			r, err := crypto.DecodeRSAPad(dhParams.EncryptedData, s.key.RSA)
			if err != nil {
				return ServerExchangeResult{}, wrapKeyNotFound(err)
			}
			b.ResetTo(r)
		}

		d, err := mt.DecodePQInnerData(b)
		if err != nil {
			return ServerExchangeResult{}, err
		}

		if innerDataDC, ok := d.(*mt.PQInnerDataDC); ok && innerDataDC.DC != s.dc {
			err := xerrors.Errorf(
				"wrong DC ID, want %d, got %d",
				s.dc, innerDataDC.DC,
			)
			return ServerExchangeResult{}, serverError(codec.CodeWrongDC, err)
		}

		innerData = mt.PQInnerData{
			Pq:          d.GetPq(),
			P:           d.GetP(),
			Q:           d.GetQ(),
			Nonce:       d.GetNonce(),
			ServerNonce: d.GetServerNonce(),
			NewNonce:    d.GetNewNonce(),
		}
	}

	dhPrime, err := s.rng.DhPrime()
	if err != nil {
		return ServerExchangeResult{}, xerrors.Errorf("generate dh_prime: %w", err)
	}

	g := 3
	a, ga, err := s.rng.GA(g, dhPrime)
	if err != nil {
		return ServerExchangeResult{}, xerrors.Errorf("generate g_a: %w", err)
	}

	data := mt.ServerDHInnerData{
		Nonce:       pqReq.Nonce,
		ServerNonce: serverNonce,
		G:           g,
		GA:          ga.Bytes(),
		DhPrime:     dhPrime.Bytes(),
		ServerTime:  int(s.clock.Now().Unix()),
	}

	b.Reset()
	if err := data.Encode(b); err != nil {
		return ServerExchangeResult{}, err
	}

	key, iv := crypto.TempAESKeys(innerData.NewNonce.BigInt(), serverNonce.BigInt())
	answer, err := crypto.EncryptExchangeAnswer(s.rand, b.Raw(), key, iv)
	if err != nil {
		return ServerExchangeResult{}, err
	}

	s.log.Debug("Sending ServerDHParamsOk", zap.Int("g", g))
	// 5. Server responds with Server_DH_Params.
	if err := s.writeUnencrypted(ctx, b, &mt.ServerDHParamsOk{
		Nonce:           pqReq.Nonce,
		ServerNonce:     serverNonce,
		EncryptedAnswer: answer,
	}); err != nil {
		return ServerExchangeResult{}, err
	}

	var clientDhParams mt.SetClientDHParamsRequest
	if err := s.readUnencrypted(ctx, b, &clientDhParams); err != nil {
		return ServerExchangeResult{}, err
	}
	s.log.Debug("Received client SetClientDHParamsRequest")

	decrypted, err := crypto.DecryptExchangeAnswer(clientDhParams.EncryptedData, key, iv)
	if err != nil {
		err = xerrors.Errorf("decrypt exchange answer: %w", err)
		return ServerExchangeResult{}, wrapKeyNotFound(err)
	}
	b.ResetTo(decrypted)

	var clientInnerData mt.ClientDHInnerData
	if err := clientInnerData.Decode(b); err != nil {
		return ServerExchangeResult{}, wrapKeyNotFound(err)
	}

	gB := big.NewInt(0).SetBytes(clientInnerData.GB)
	var authKey crypto.Key
	if !crypto.FillBytes(big.NewInt(0).Exp(gB, a, dhPrime), authKey[:]) {
		err := xerrors.New("auth_key is too big")
		return ServerExchangeResult{}, wrapKeyNotFound(err)
	}

	// DH key exchange complete
	// 8. Server responds in one of three ways:
	// dh_gen_ok#3bcbf734 nonce:int128 server_nonce:int128
	// 	new_nonce_hash1:int128 = Set_client_DH_params_answer;
	s.log.Debug("Sending DhGenOk")
	if err := s.writeUnencrypted(ctx, b, &mt.DhGenOk{
		Nonce:         pqReq.Nonce,
		ServerNonce:   serverNonce,
		NewNonceHash1: crypto.NonceHash1(innerData.NewNonce, authKey),
	}); err != nil {
		return ServerExchangeResult{}, err
	}

	serverSalt := crypto.ServerSalt(innerData.NewNonce, serverNonce)
	return ServerExchangeResult{
		Key:        authKey.WithID(),
		ServerSalt: serverSalt,
	}, nil
}
