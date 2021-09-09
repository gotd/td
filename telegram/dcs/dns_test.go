package dcs

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/dns/dnsmessage"
)

func Test_DNSConfig(t *testing.T) {
	a := require.New(t)

	cfg, err := DNSConfig(dnsmessage.TXTResource{
		TXT: []string{
			"LcmEoukF2bVjKwz3E+J9BsDdL+rv9lGqLQWIGXrWACT2ESk5xuOpA6Cz6klKRbhbwSiHOd2zC5PiR57j/OJHPpj4i+tw==",
			"umjjLFLpOKtPeW9zHLq2ypbMzg/zkqvPhvhr0bxrLZlgPQ04l2GpO/4qZgAx3tk3BDHbY6/gmG1e8eaFBq3YSqR5SZ5hQ1Cm5f4/" +
				"o67GYcPJClaf1TiHq3wVfsQ5OLnyJRw9A2ZfUfzIXxoSklPJrVdF/4hM1ZdUE0eWDAbmYf7JCeao8ecVVwKndd4CZHZS9wyf1T7DIUh95VpQ" +
				"sn2klLPA6gA/2YNXOh9gITvjZrKuXLwwh9hBHhPvxv",
		},
	})
	a.NoError(err)
	a.Equal(1565541126, cfg.Expires)
	a.Equal(1562949126, cfg.Date)
	a.Len(cfg.Rules, 1)

	rule := cfg.Rules[0]
	a.Equal(2, rule.DCID)
}
