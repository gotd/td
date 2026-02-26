package pool

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/mtproto"
)

func TestSyncSessionOptions(t *testing.T) {
	a := require.New(t)

	session := NewSyncSession(Session{
		DC:      2,
		AuthKey: crypto.Key{1}.WithID(),
		Salt:    42,
	})
	opts, data := session.Options(mtproto.Options{})

	a.Equal(2, data.DC)
	a.Equal(int64(42), opts.Salt)
	a.Equal(data.AuthKey, opts.Key)
	a.True(opts.PermKey.Zero())
}

func TestSyncSessionOptionsPFS(t *testing.T) {
	a := require.New(t)

	session := NewSyncSession(Session{
		DC:      2,
		AuthKey: crypto.Key{1}.WithID(),
		Salt:    42,
	})
	opts, data := session.Options(mtproto.Options{
		EnablePFS: true,
	})

	// PFS mode must move persisted key into PermKey and force runtime Key=zero.
	a.Equal(2, data.DC)
	a.Equal(int64(42), opts.Salt)
	a.True(opts.Key.Zero())
	a.Equal(data.AuthKey, opts.PermKey)
}
