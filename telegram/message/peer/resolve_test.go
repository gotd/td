package peer

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func resolver(t *testing.T, expectedDomain string, expected tg.InputPeerClass) Resolver {
	return &mockResolver{
		domain: expectedDomain,
		peer:   expected,
		t:      t,
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
		{"Bad", "_gotd_test", true},
		{"Bad", "gotd_test_", true},
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

					p, err := Resolve(
						resolver(t, tt.domain, expected),
						fmt.Sprintf(format.fmt, tt.domain),
					)(context.Background())
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

func Test_cleanupPhone(t *testing.T) {
	tests := []struct {
		phone string
		want  string
	}{
		{"+13115552368", "13115552368"},
		{"+1 (311) 555-0123", "13115550123"},
		{"+1 311 555-6162", "13115556162"},
		{"13115556162", "13115556162"},
		{"123gotd_test", "123"},
	}
	for _, tt := range tests {
		t.Run(tt.phone, func(t *testing.T) {
			r := cleanupPhone(tt.phone)
			require.Equal(t, tt.want, r)
			_, err := strconv.ParseInt(r, 10, 64)
			require.NoError(t, err)
		})
	}
}
