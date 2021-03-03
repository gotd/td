package message

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

func resolver(t *testing.T, expectedDomain string, expected tg.InputPeerClass) peer.ResolverFunc {
	return func(ctx context.Context, domain string) (tg.InputPeerClass, error) {
		if domain != expectedDomain {
			err := fmt.Errorf("expected domain %q, got %q", expectedDomain, domain)
			t.Error(err)
			return nil, err
		}
		return expected, nil
	}
}

func TestSender_Resolve(t *testing.T) {
	formats := []struct {
		fmt     string
		wantErr bool
	}{
		{`%s`, false},
		{`@%s`, false},
		{`t.me/%s`, false},
		{`t.me/%s/`, false},
		{`https://t.me/%s`, false},
		{`https://t.me/%s/`, false},
		{`tg:resolve?domain=%s`, false},
		{`tg://resolve?domain=%s`, false},

		{`https://t.co/%s`, true},
		{`rt://resolve?domain=%s`, true},
	}

	tests := []struct {
		name    string
		domain  string
		wantErr bool
	}{
		{"Good", "telegram", false},
		{"Good with numbers", "telegram123", false},
		{"Good with _", "telegram_test", false},
		{"Good with numbers and _", "telegram_test123", false},
		{"Bad", "", true},
		{"Bad", "gotd", true},
		{"Bad", "_gotd_test", true},
		{"Bad", "gotd_test_", true},
		{"Bad", "123gotd_test", true},
		{"Bad", "_gotd_test123", true},
		{"Bad", "gotd.test", true},
		{"Bad", "gotd/test", true},
	}

	for _, format := range formats {
		t.Run(format.fmt, func(t *testing.T) {
			for _, tt := range tests {
				name := tt.name
				if tt.wantErr {
					name = fmt.Sprintf("%s (%q)", tt.name, tt.domain)
				}

				expected := &tg.InputPeerUser{
					UserID:     1,
					AccessHash: 10,
				}
				t.Run(name, func(t *testing.T) {
					a := require.New(t)
					s := &Sender{
						resolver: resolver(t, tt.domain, expected),
					}

					p, err := s.Resolve(fmt.Sprintf(format.fmt, tt.domain)).AsInputPeer(context.Background())
					if tt.wantErr || format.wantErr {
						a.Error(err)
						return
					}
					a.NoError(err)
					a.Equal(expected, p)
				})
			}
		})
	}
}
