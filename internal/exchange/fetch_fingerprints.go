package exchange

import (
	"context"
	"crypto/rand"
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/internal/crypto"
	"github.com/gotd/td/internal/mt"
	"github.com/gotd/td/internal/proto"
	"github.com/gotd/td/transport"
)

// FetchFingerprints gets fingerprints using given connection.
func FetchFingerprints(ctx context.Context, conn transport.Conn) ([]int64, error) {
	nonce, err := crypto.RandInt128(rand.Reader)
	if err != nil {
		return nil, xerrors.Errorf("generate nonce: %w", err)
	}

	w := unencryptedWriter{
		clock:   clock.System,
		conn:    conn,
		timeout: 30 * time.Second,
		input:   proto.MessageServerResponse,
		output:  proto.MessageFromClient,
	}

	var buf bin.Buffer
	if err := w.writeUnencrypted(ctx, &buf, &mt.ReqPqMultiRequest{Nonce: nonce}); err != nil {
		return nil, xerrors.Errorf("write ReqPqMultiRequest: %w", err)
	}

	var res mt.ResPQ
	if err := w.readUnencrypted(ctx, &buf, &res); err != nil {
		return nil, xerrors.Errorf("read ResPQ response: %w", err)
	}
	if res.Nonce != nonce {
		return nil, xerrors.New("ResPQ nonce mismatch")
	}
	return res.ServerPublicKeyFingerprints, nil
}
