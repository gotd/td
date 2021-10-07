package session

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/internal/crypto"
)

func repeat(i int) []byte {
	return bytes.Repeat([]byte{'a'}, i)
}

var (
	testKey = repeat(256)
	ipv4    = append([]byte("\x02\xc0\xa8\x00\x01\x01\xbb"), testKey...)
	ipv6    = append([]byte("\x02 \x01\r\xb8\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\xbb"), testKey...)
)

func TestStringSession(t *testing.T) {
	var key crypto.Key
	copy(key[:], testKey)
	authKey := key.WithID()

	based := func(data []byte) string {
		return "1" + base64.URLEncoding.EncodeToString(data)
	}

	tests := []struct {
		name    string
		hx      string
		want    *Data
		wantErr bool
	}{
		{"IPv4", based(ipv4), &Data{
			DC:        2,
			Addr:      "192.168.0.1:443",
			AuthKey:   authKey.Value[:],
			AuthKeyID: authKey.ID[:],
		}, false},
		{"IPv6", based(ipv6),
			&Data{
				DC:        2,
				Addr:      "[2001:db8::]:443",
				AuthKey:   authKey.Value[:],
				AuthKeyID: authKey.ID[:],
			}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.hx)
			got, err := TelethonSession(tt.hx)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_decodeStringSession(t *testing.T) {
	var key crypto.Key
	copy(key[:], testKey)
	authKey := key.WithID()

	tests := []struct {
		name    string
		data    []byte
		want    *Data
		wantErr bool
	}{
		{"TooSmall", repeat(1), nil, true},
		{"InvalidLength", repeat(267), nil, true},
		{"TooBig", repeat(276), nil, true},
		{"IPv4", ipv4, &Data{
			DC:        2,
			Addr:      "192.168.0.1:443",
			AuthKey:   authKey.Value[:],
			AuthKeyID: authKey.ID[:],
		}, false},
		{"IPv6", ipv6, &Data{
			DC:        2,
			Addr:      "[2001:db8::]:443",
			AuthKey:   authKey.Value[:],
			AuthKeyID: authKey.ID[:],
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeStringSession(tt.data)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
