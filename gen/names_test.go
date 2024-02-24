package gen

import "testing"

func TestPascal(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "Id",
			expected: "ID",
		},
		{
			name:     "user_id",
			expected: "UserID",
		},
		{
			name:     "cdnConfig",
			expected: "CDNConfig",
		},
		{
			name:     "cdn_1_config",
			expected: "CDN1Config",
		},
		{
			name:     "p2pB2B",
			expected: "P2PB2B",
		},
		{
			name:     "md5Checksum",
			expected: "MD5Checksum",
		},
		{
			name:     "user_ids",
			expected: "UserIDs",
		},
		{
			name:     "UserIDs",
			expected: "UserIDs",
		},
		{
			name:     "tcpo_only",
			expected: "TCPObfuscatedOnly",
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			res := pascal(test.name)
			if res != test.expected {
				t.Fatalf("mismatch; got: %s; expected: %s", res, test.expected)
			}
		})
	}
}

func TestCamel(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "user_id",
			expected: "userID",
		},
		{
			name:     "full_name",
			expected: "fullName",
		},
		{
			name:     "full-admin",
			expected: "fullAdmin",
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			res := camel(test.name)
			if res != test.expected {
				t.Fatalf("mismatch; got: %s; expected: %s", res, test.expected)
			}
		})
	}
}
