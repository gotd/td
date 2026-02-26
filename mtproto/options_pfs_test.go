package mtproto

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/crypto"
)

func TestOptionsSetDefaultsPFS(t *testing.T) {
	t.Run("DefaultTTL", func(t *testing.T) {
		a := require.New(t)
		opt := Options{
			EnablePFS: true,
		}
		opt.setDefaults()
		a.Equal(defaultTempKeyTTL, opt.TempKeyTTL)
	})

	t.Run("MinTTL", func(t *testing.T) {
		a := require.New(t)
		opt := Options{
			EnablePFS:  true,
			TempKeyTTL: 1,
		}
		opt.setDefaults()
		a.Equal(minTempKeyTTL, opt.TempKeyTTL)
	})

	t.Run("PermKeyFallback", func(t *testing.T) {
		a := require.New(t)
		key := crypto.Key{1}.WithID()
		opt := Options{
			EnablePFS: true,
			Key:       key,
		}
		// Compatibility: old callers restore only Key, so defaults must map it
		// into PermKey in PFS mode.
		opt.setDefaults()
		a.Equal(key, opt.PermKey)
	})
}
