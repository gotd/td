package faketls

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/internal/mtproxy/faketls/tlstypes"
)

func generate32(t *testing.T) [32]byte {
	result := [32]byte{}
	_, err := io.ReadFull(rand.Reader, result[:])
	require.NoError(t, err)
	return result
}

func TestTLS(t *testing.T) {
	a := require.New(t)
	secret := generate32(t)
	sessionID := generate32(t)

	b := bytes.NewBuffer(nil)

	clientRandom, err := writeClientHello(b, sessionID, secret[:])
	a.NoError(err)

	helloRecord, err := tlstypes.ReadRecord(b)
	a.NoError(err)

	buf := &bytes.Buffer{}
	helloRecord.Data.WriteBytes(buf)

	clientHello, err := tlstypes.ParseClientHello(buf.Bytes())
	a.NoError(err)
	a.Equal(clientRandom, clientHello.Random)

	digest := clientHello.Digest(secret[:])
	for i := 0; i < len(digest)-4; i++ {
		a.Zero(digest[i])
	}

	timestamp := int64(binary.LittleEndian.Uint32(digest[len(digest)-4:]))
	createdAt := time.Unix(timestamp, 0)
	a.WithinDuration(time.Now(), createdAt, 5*time.Second)

	serverHello := tlstypes.NewServerHello(clientHello).WelcomePacket(secret[:])
	err = readServerHello(bytes.NewReader(serverHello), clientRandom, secret[:])
	a.NoError(err)
}
