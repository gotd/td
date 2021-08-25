package tdesktop

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_createLocalKey(t *testing.T) {
	tests := []struct {
		name       string
		pass, salt []byte
		output     string
	}{
		{
			"NoPasscode",
			nil,
			[]byte("salt"),
			"bd73811ec41b37e1bbe6484bbb42cb775d3aa83a453a80" +
				"380b35f6773319b789a92ca49a8f4607ce412e7955e5916c8" +
				"8047b237e60e4bcc28d0a21628f07cefad449f996ac42ab30" +
				"25b80005f9d5d75c12e4927782b12ed77ce15c96e1a44a6bf" +
				"65dfa67e8228d7351b12336692223cee72d697f226cfee229" +
				"54196856100d7e1cfc70b0b04deb30190502f3438e06530e1" +
				"8253d4c3d87daa1d1a0ad27e537f49baf6835cc6b2cf701e7" +
				"fb8a457d04bd092372c9fc5d9b4cc8be2a62a979333eb736a" +
				"2e72b6b6e8da385117092e9a4eb0797098e9f2f156f0cdbcd" +
				"ea5c27d5e2decf1bb383e7b8568ed1f384bf84de414a07595" +
				"6498c6903d4ac6612c43b7eea",
		},
		{
			"pass",
			[]byte("pass"),
			[]byte("salt"),
			"54f00fbe5fbd1ddbc42f290e892032f780dff189d759fc4" +
				"5f1bb14d03db1a6a37a7ba27402dd53a3429657afc293ff26f" +
				"15b4df1351502386844e0ab213f4662ffa7dd5e60b0e06abfd" +
				"5ee0d6b7a266a86cbf3aa5edaa92ab3e992aa20a31becf860f" +
				"b48689310144c6d1a9d98f90b84675b7fe00c41782e940db04" +
				"f6bb84babdd350f1d45fc3bb2073d42f36ba47bdfe93d4c969" +
				"9291b62d0e7bc01765886a9475f412420d9609903f6a654425" +
				"63de1dbea16f6c8d758ed43a6ccc3cb7fe9c5f9a0df8285deb" +
				"db25e1d1dd39ecd584a41b161b39ff1becd42735bc44118d39" +
				"d71aa4ffce8e8559a6b901a17379620e1fadea76cb51a4ff18" +
				"b7b962ee6076375",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)
			expected, err := hex.DecodeString(tt.output)
			a.NoError(err)

			k := createLocalKey(tt.pass, tt.salt)
			a.Equal(expected, k[:])
		})
	}
}
