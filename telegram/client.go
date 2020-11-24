package telegram

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/ernado/td/bin"
	"github.com/ernado/td/internal/crypto"
	"github.com/ernado/td/internal/mt"
	"github.com/ernado/td/internal/proto"
)

// Client represents a MTProto client to Telegram.
type Client struct {
	conn  net.Conn
	clock func() time.Time

	rsaPublicKeys []*rsa.PublicKey
}

const defaultTimeout = time.Second * 10

func (c Client) startIntermediateMode(deadline time.Time) error {
	if err := c.conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}
	if _, err := c.conn.Write(proto.IntermediateClientStart); err != nil {
		return fmt.Errorf("failed to write start: %w", err)
	}
	if err := c.conn.SetDeadline(time.Time{}); err != nil {
		return fmt.Errorf("failed to reset connection deadline: %w", err)
	}
	return nil
}

func (c Client) resetDeadline() error {
	return c.conn.SetDeadline(time.Time{})
}

func (c Client) deadline(ctx context.Context) time.Time {
	if deadline, ok := ctx.Deadline(); ok {
		return deadline
	}
	return c.clock().Add(defaultTimeout)
}

func (c Client) newUnencryptedMessage(payload bin.Encoder, b *bin.Buffer) error {
	b.Reset()
	if err := payload.Encode(b); err != nil {
		return err
	}
	msg := proto.UnencryptedMessage{
		MessageID:   crypto.NewMessageID(c.clock(), crypto.ModeClient),
		MessageData: append([]byte(nil), b.Buf...),
	}
	b.Reset()
	return msg.Encode(b)
}

// CreateAuthKey generates new authorization key.
func (c Client) CreateAuthKey(ctx context.Context) error {
	// NOTE: Currently WIP.

	if err := c.conn.SetDeadline(c.deadline(ctx)); err != nil {
		return err
	}
	defer func() { _ = c.resetDeadline() }()

	// 1. DH exchange initiation.
	nonce, err := crypto.RandInt128(rand.Reader)
	if err != nil {
		return err
	}
	b := new(bin.Buffer)
	if err := c.newUnencryptedMessage(&mt.ReqPqMulti{Nonce: nonce}, b); err != nil {
		return err
	}
	if err := proto.WriteIntermediate(c.conn, b); err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}
	b.Reset()

	// 2. Server sends response of the form
	// resPQ#05162463 nonce:int128 server_nonce:int128 pq:string server_public_key_fingerprints:Vector long = ResPQ;
	if err := proto.ReadIntermediate(c.conn, b); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	msg := proto.UnencryptedMessage{}
	if err := msg.Decode(b); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	b.ResetTo(msg.MessageData)
	res := mt.ResPQ{}
	if err := res.Decode(b); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	if res.Nonce != nonce {
		return errors.New("nonce mismatch")
	}

	// Selecting first public key that match fingerprint.
	var selectedPubKey *rsa.PublicKey
Loop:
	for _, fingerprint := range res.ServerPublicKeyFingerprints {
		for _, key := range c.rsaPublicKeys {
			if fingerprint == proto.RSAFingerprint(key) {
				selectedPubKey = key
				break Loop
			}
		}
	}
	if selectedPubKey == nil {
		return errors.New("unable to select public key")
	}

	// The pq is a representation of a natural number (in binary big endian format).
	// SetBytes is also big endian.
	pq := big.NewInt(0).SetBytes(res.Pq)
	// Normally pq is less than or equal to 2^63-1.
	pqMax := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(63), nil)
	if pq.Cmp(pqMax) > 0 {
		return errors.New("server provided bad pq")
	}

	// 3. Client decomposes pq into prime factors such that p < q.
	// Performing proof of work.
	p, q, err := crypto.DecomposePQ(pq, rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to decompose pq: %w", err)
	}

	// 4. Client sends query to server.
	// req_DH_params#d712e4be nonce:int128 server_nonce:int128 p:string q:string
	//   public_key_fingerprint:long encrypted_data:string = Server_DH_Params
	newNonce, err := crypto.RandInt256(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate new nonce: %w", err)
	}
	pqInnerData := &mt.PQInnerDataConst{
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
	encryptedData, err := crypto.EncryptHashed(b.Buf, selectedPubKey, rand.Reader)
	if err != nil {
		return err
	}
	reqDHParams := &mt.ReqDHParams{
		Nonce:                nonce,
		ServerNonce:          res.ServerNonce,
		P:                    p.Bytes(),
		Q:                    q.Bytes(),
		PublicKeyFingerprint: proto.RSAFingerprint(selectedPubKey),
		EncryptedData:        encryptedData,
	}
	if err := c.newUnencryptedMessage(reqDHParams, b); err != nil {
		return err
	}
	if err := proto.WriteIntermediate(c.conn, b); err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}
	b.Reset()

	// 5. Server responds with Server_DH_Params.
	if err := proto.ReadIntermediate(c.conn, b); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if err := msg.Decode(b); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	b.ResetTo(msg.MessageData)
	dhParams, err := mt.DecodeServerDHParams(b)
	if err != nil {
		return fmt.Errorf("failed to decode Server_DH_Params: %w", err)
	}
	switch p := dhParams.(type) {
	case *mt.ServerDHParamsOk:
		// Success.
		if p.Nonce != nonce {
			return errors.New("nonce mismatch")
		}
		// TODO: Decode inner data.
		// server_DH_inner_data#b5890dba nonce:int128 server_nonce:int128 g:int dh_prime:string g_a:string server_time:int = Server_DH_inner_data;
	case *mt.ServerDHParamsFail:
		return errors.New("server respond with server_DH_params_fail")
	default:
		return fmt.Errorf("unknown ")
	}

	// TODO: Complete.
	return nil
}

type Options struct {
	Dialer     *net.Dialer
	PublicKeys []*rsa.PublicKey
	Network    string
	Addr       string
}

func Dial(ctx context.Context, opt Options) (*Client, error) {
	if opt.Dialer == nil {
		opt.Dialer = &net.Dialer{}
	}
	if opt.Network == "" {
		opt.Network = "tcp"
	}
	conn, err := opt.Dialer.DialContext(ctx, "tcp", opt.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}
	client := &Client{
		conn:          conn,
		clock:         time.Now,
		rsaPublicKeys: opt.PublicKeys,
	}
	return client, nil
}

// Connect establishes connection in intermediate mode.
func (c *Client) Connect(ctx context.Context) error {
	deadline := c.clock().Add(defaultTimeout)
	if ctxDeadline, ok := ctx.Deadline(); ok {
		deadline = ctxDeadline
	}
	if err := c.startIntermediateMode(deadline); err != nil {
		return err
	}
	return nil
}
