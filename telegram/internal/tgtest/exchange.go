package tgtest

import (
	"math/big"
	"net"
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
)

// nolint:gocognit,gocyclo // TODO(tdakkota): simplify
func (s *Server) exchange(conn net.Conn) (crypto.AuthKey, error) {
	// 1. Client sends query to server
	//
	// req_pq_multi#be7e8ef1 nonce:int128 = ResPQ;
	var pqReq mt.ReqPqMulti
	if err := s.readUnencrypted(conn, &pqReq); err != nil {
		return crypto.AuthKey{}, err
	}

	serverNonce, err := crypto.RandInt128(s.cipher.Rand())
	if err != nil {
		return crypto.AuthKey{}, xerrors.Errorf("failed to generate server nonce: %w", err)
	}

	// 2. Server sends response of the form
	//
	// resPQ#05162463 nonce:int128 server_nonce:int128 pq:string server_public_key_fingerprints:Vector long = ResPQ;
	pq, err := s.pq()
	if err != nil {
		return crypto.AuthKey{}, xerrors.Errorf("failed to generate pq: %w", err)
	}

	if err := s.writeUnencrypted(conn, &mt.ResPQ{
		Pq:          pq.Bytes(),
		Nonce:       pqReq.Nonce,
		ServerNonce: serverNonce,
		ServerPublicKeyFingerprints: []int64{
			crypto.RSAFingerprint(s.Key()),
		},
	}); err != nil {
		return crypto.AuthKey{}, err
	}

	// TODO(ernado): make actual crypto here
	// 4. Client sends query to server
	//
	// req_DH_params#d712e4be nonce:int128 server_nonce:int128 p:string
	//  q:string public_key_fingerprint:long encrypted_data:string = Server_DH_Params
	var dhParams mt.ReqDHParams
	if err := s.readUnencrypted(conn, &dhParams); err != nil {
		return crypto.AuthKey{}, err
	}

	var b bin.Buffer
	b.Put(crypto.RSADecryptHashed(dhParams.EncryptedData, s.key))

	var innerData mt.PQInnerDataConst
	err = innerData.Decode(&b)
	if err != nil {
		return crypto.AuthKey{}, err
	}

	dhPrime, err := s.dhPrime()
	if err != nil {
		return crypto.AuthKey{}, xerrors.Errorf("failed to generate dh_prime: %w", err)
	}

	g := 2
	a, ga, err := s.ga(big.NewInt(int64(g)), dhPrime)
	if err != nil {
		return crypto.AuthKey{}, xerrors.Errorf("failed to generate g_a: %w", err)
	}

	data := mt.ServerDHInnerData{
		Nonce:       pqReq.Nonce,
		ServerNonce: serverNonce,
		G:           g,
		GA:          ga.Bytes(),
		DhPrime:     dhPrime.Bytes(),
		ServerTime:  int(time.Now().Unix()),
	}

	b.Reset()
	err = data.Encode(&b)
	if err != nil {
		return crypto.AuthKey{}, err
	}

	answer, err := s.encryptedExchangeAnswer(b.Raw(), innerData.NewNonce, serverNonce)
	if err != nil {
		return crypto.AuthKey{}, err
	}

	// 5. Server responds with Server_DH_Params.
	if err := s.writeUnencrypted(conn, &mt.ServerDHParamsOk{
		Nonce:           pqReq.Nonce,
		ServerNonce:     serverNonce,
		EncryptedAnswer: answer,
	}); err != nil {
		return crypto.AuthKey{}, err
	}

	var clientDhParams mt.SetClientDHParams
	if err := s.readUnencrypted(conn, &clientDhParams); err != nil {
		return crypto.AuthKey{}, err
	}

	decrypted, err := s.decryptedExchangeAnswer(clientDhParams.EncryptedData, innerData.NewNonce, serverNonce)
	if err != nil {
		return crypto.AuthKey{}, err
	}
	b.Reset()
	b.Put(decrypted)

	var clientInnerData mt.ClientDHInnerData
	err = clientInnerData.Decode(&b)
	if err != nil {
		return crypto.AuthKey{}, err
	}

	gB := big.NewInt(0).SetBytes(clientInnerData.GB)
	authKey := big.NewInt(0).Exp(gB, a, dhPrime).Bytes()

	// 8. Server responds in one of three ways:
	if err := s.writeUnencrypted(conn, &mt.DhGenOk{
		Nonce:         pqReq.Nonce,
		ServerNonce:   serverNonce,
		NewNonceHash1: s.getNonceHash1(authKey, innerData.NewNonce[:]),
	}); err != nil {
		return crypto.AuthKey{}, err
	}

	var result crypto.AuthKey
	copy(result[:], authKey)
	return result, nil
}
