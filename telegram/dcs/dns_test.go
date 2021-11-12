package dcs

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func getDNSTestData() []string {
	return []string{
		"LcmEoukF2bVjKwz3E+J9BsDdL+rv9lGqLQWIGXrWACT2ESk5xuOpA6Cz6klKRbhbwSiHOd2zC5PiR57j/OJHPpj4i+tw==",
		"umjjLFLpOKtPeW9zHLq2ypbMzg/zkqvPhvhr0bxrLZlgPQ04l2GpO/4qZgAx3tk3BDHbY6/gmG1e8eaFBq3YSqR5SZ5hQ1" +
			"Cm5f4/o67GYcPJClaf1TiHq3wVfsQ5OLnyJRw9A2ZfUfzIXxoSklPJrVdF/4hM1ZdUE0eWDAbmYf7JCeao8ecVVwKn" +
			"dd4CZHZS9wyf1T7DIUh95VpQsn2klLPA6gA/2YNXOh9gITvjZrKuXLwwh9hBHhPvxv",
	}
}

func Test_DNSConfig(t *testing.T) {
	t.Run("Good", func(t *testing.T) {
		a := require.New(t)

		cfg, err := DNSConfig(getDNSTestData())
		a.NoError(err)
		a.Equal(1565541126, cfg.Expires)
		a.Equal(1562949126, cfg.Date)
		a.Len(cfg.Rules, 1)

		rule := cfg.Rules[0]
		a.Equal(2, rule.DCID)
	})
	t.Run("Bad", func(t *testing.T) {
		tests := []struct {
			name string
			txt  []string
		}{
			{"Empty", nil},
			{"InvalidLen", getDNSTestData()[:1]},
			{"InvalidBas64", func() []string {
				r := getDNSTestData()
				first := r[0]
				r[0] = "#" + first[1:]
				return r
			}()},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := DNSConfig(tt.txt)
				require.Error(t, err)
			})
		}
	})
}

func BenchmarkDNSConfig(b *testing.B) {
	message := getDNSTestData()

	b.ResetTimer()
	b.ReportAllocs()

	var (
		err     error
		cfgSink tg.HelpConfigSimple
	)
	for i := 0; i < b.N; i++ {
		cfgSink, err = DNSConfig(message)
		if cfgSink.Zero() || err != nil {
			b.Fatal(err)
		}
	}
}
