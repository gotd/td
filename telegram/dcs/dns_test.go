package dcs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testTXTResponse() []string {
	return []string{
		"LcmEoukF2bVjKwz3E+J9BsDdL+rv9lGqLQWIGXrWACT2ESk5xuOpA6Cz6klKRbhbwSiHOd2zC5PiR57j/OJHPpj4i+tw==",
		"umjjLFLpOKtPeW9zHLq2ypbMzg/zkqvPhvhr0bxrLZlgPQ04l2GpO/4qZgAx3tk3BDHbY6/gmG1e8eaFBq3YSqR5SZ5hQ1Cm5f4/" +
			"o67GYcPJClaf1TiHq3wVfsQ5OLnyJRw9A2ZfUfzIXxoSklPJrVdF/4hM1ZdUE0eWDAbmYf7JCeao8ecVVwKndd4CZHZS9wyf1T7DIUh95VpQ" +
			"sn2klLPA6gA/2YNXOh9gITvjZrKuXLwwh9hBHhPvxv",
	}
}

func Test_ParseDNSConfig(t *testing.T) {
	t.Run("Good", func(t *testing.T) {
		a := require.New(t)

		cfg, err := ParseDNSConfig(testTXTResponse())
		a.NoError(err)
		a.Equal(1565541126, cfg.Expires)
		a.Equal(1562949126, cfg.Date)
		a.Len(cfg.Rules, 1)

		rule := cfg.Rules[0]
		a.Equal(2, rule.DCID)
	})

	t.Run("Bad", func(t *testing.T) {
		tests := []struct {
			name  string
			input []string
		}{
			{"Empty", nil},
			{"InvalidHash", func() []string {
				r := testTXTResponse()
				first := r[0]
				r[0] = string(first[0]+1) + first[1:]
				return r
			}()},
			{"InvalidBase64", func() []string {
				r := testTXTResponse()
				first := r[0]
				r[0] = string('#') + first[1:]
				return r
			}()},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := ParseDNSConfig(tt.input)
				require.Error(t, err)
			})
		}
	})
}

func BenchmarkDNSConfig(b *testing.B) {
	message := testTXTResponse()

	b.ResetTimer()
	b.ReportAllocs()

	var (
		err     error
		cfgSink DNSConfig
	)
	for i := 0; i < b.N; i++ {
		cfgSink, err = ParseDNSConfig(message)
		if cfgSink.Date == 0 || err != nil {
			b.Fatal(err)
		}
	}
}

func TestDNSConfig_Options(t *testing.T) {
	a := require.New(t)

	cfg, err := ParseDNSConfig(testTXTResponse())
	a.NoError(err)

	options := cfg.Options()
	a.Len(options, 1)
	option := options[0]
	a.Equal(2, option.ID)
	a.Equal(14544, option.Port)
	a.Equal("98.210.59.139", option.IPAddress)
	a.True(option.TCPObfuscatedOnly)
}
