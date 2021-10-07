package tdesktop

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/internal/crypto"
)

func Test_createLegacyLocalKey(t *testing.T) {
	tests := []struct {
		name       string
		pass, salt []byte
		output     string
	}{
		{
			"NoPasscode",
			nil,
			[]byte("salt"),
			"49c9b4503692e6d97dd4dc009d25f0c3ba18f2c24ba2247" +
				"63e9e9b7a105defea90a7af133c8877199219be40aad81df80" +
				"3785f07b4ad88cc4a03be6946a3aca1fc74d6bbb74d39d975c" +
				"59cda120226493b4937ca99c933423ee15352c8e76efc9c3dc" +
				"b4f5d4ee9f123ee5e339ccfe3c84909290a002bc91d29fa27f" +
				"66fb736d22bd6d4ab7a5020d31dcd0d491042d78522f2470a4" +
				"4281cc9315856e1528d5abbe1d78573230d73516eedce9598c" +
				"dee1052f73e154fce79d9934a66e1b52b1d598861648a4f9d9" +
				"5a958a5f527c896db63ff7e1dae0db16c66c36ba984faf65f3" +
				"36fdb4f7efcbaee7f89bf634ef084bbf6e46d91f8ceaf4052e" +
				"9ea20f49bf243dc",
		},
		{
			"pass",
			[]byte("pass"),
			[]byte("salt"),
			"b08653719bf59a6a7c8eb1abae9c267e6e0252a9ea54683" +
				"806d093c2f1dff9cea4341b3728bca217389026afe6c7b69eb" +
				"9affc6e3ced50b07e0168fc4ad2cef468f06def70cc932b7e6" +
				"024f3c92bf3f650ed49df4460b0fbd30358c57c4db14ac7ddb" +
				"755dff9d1b0b7c664e11bd3460f0f772a9ac1afd880d3be01c" +
				"8b39ccd44e96248226cbe5623436abb0ef26071eafbc7b8cfb" +
				"1db72c982dd7a61baa2669ada459e2c6d67ad5e7c1445ed48e" +
				"0b8e3ec4fbeb5126bec2175508acb9e0e1f9aa2f7ea888e519" +
				"85b410e41da33fc38d765cb00ace54860069edfc8a35c9650c" +
				"754989defefc785772fd7eb017b1ef351cf3abcc839ce2c995" +
				"5981d555bbaefad",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)
			expected, err := hex.DecodeString(tt.output)
			a.NoError(err)

			k := createLegacyLocalKey(tt.pass, tt.salt)
			a.Equal(expected, k[:])
		})
	}
}

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

func Test_cryptLocal(t *testing.T) {
	key := crypto.Key{
		0xe1, 0x6f, 0x98, 0xd9, 0xf0, 0x7e, 0x58, 0x58,
		0xda, 0x7a, 0xf3, 0xf9, 0xd0, 0x04, 0x1a, 0xa0,
		0x11, 0xd2, 0x33, 0xd1, 0x7e, 0x58, 0xa9, 0x05,
		0xd0, 0x82, 0x11, 0x97, 0xa5, 0x6b, 0xd0, 0x69,
		0x3d, 0x86, 0x79, 0xff, 0xef, 0x63, 0x20, 0xec,
		0xbf, 0x56, 0xa1, 0xf6, 0x12, 0x68, 0xd1, 0xd8,
		0xb8, 0x4d, 0x16, 0x15, 0x46, 0xe7, 0x1a, 0x4b,
		0xc3, 0x8d, 0x7a, 0x25, 0x59, 0x7a, 0xee, 0xef,
		0x55, 0xed, 0x01, 0x65, 0x55, 0xf1, 0x66, 0xc5,
		0xe0, 0x65, 0x5f, 0x26, 0xee, 0x40, 0x1c, 0xee,
		0x53, 0x4e, 0xd4, 0xa2, 0x67, 0xc7, 0x7a, 0xaf,
		0x23, 0x90, 0x31, 0x2b, 0xd2, 0xdd, 0xb5, 0xa9,
		0x40, 0xb5, 0xd1, 0x1d, 0x5e, 0x6c, 0xbf, 0x6f,
		0xe4, 0xb8, 0x66, 0xf3, 0x5b, 0xac, 0x1c, 0x7c,
		0xb0, 0x0d, 0x16, 0x27, 0xa3, 0x97, 0xa0, 0xdc,
		0x2b, 0xc4, 0x18, 0x8c, 0xf1, 0xe3, 0x5c, 0x6f,
		0x9f, 0xa2, 0xb2, 0x05, 0x87, 0x03, 0x70, 0xec,
		0xe6, 0x12, 0x7c, 0x36, 0x17, 0xfc, 0xc2, 0x5c,
		0x6c, 0x2f, 0xcc, 0x0f, 0x4f, 0x2c, 0xa5, 0xcc,
		0x08, 0xa5, 0x4e, 0x8b, 0xb0, 0xba, 0xb9, 0x29,
		0x6c, 0x02, 0x79, 0xb2, 0x2d, 0x73, 0xbd, 0x8b,
		0x1e, 0x9a, 0x49, 0x11, 0x9d, 0xa8, 0x88, 0xe8,
		0xb9, 0x1c, 0x32, 0x67, 0x4d, 0xf2, 0x2c, 0xa4,
		0x72, 0xa5, 0x0a, 0xdd, 0x60, 0xe3, 0xb2, 0x01,
		0x52, 0x38, 0x8e, 0xe9, 0x7b, 0x96, 0xa4, 0xbb,
		0x24, 0x0a, 0x13, 0x8f, 0x79, 0x23, 0xcc, 0x8b,
		0x82, 0x1a, 0xfb, 0xaa, 0x1e, 0xf3, 0xbe, 0x51,
		0xaa, 0xa3, 0x14, 0x83, 0x25, 0x11, 0x0e, 0xcc,
		0x7e, 0x99, 0xba, 0x37, 0x60, 0x4b, 0x72, 0x69,
		0x8e, 0xfe, 0xa4, 0xed, 0x56, 0xbf, 0xe5, 0x54,
		0x45, 0xe7, 0x2e, 0x2b, 0x55, 0x6b, 0x13, 0xe3,
		0xca, 0xe9, 0xe6, 0xa9, 0xb9, 0x89, 0xb0, 0x72,
	}

	t.Run("PaddingCheck", func(t *testing.T) {
		a := require.New(t)
		_, err := encryptLocal([]byte{1}, key)
		a.Error(err)
		_, err = decryptLocal([]byte{1}, key)
		a.Error(err)
	})
	t.Run("DecryptEncryptDecrypt", func(t *testing.T) {
		expectEncrypted := []uint8{
			0x0a, 0x9e, 0x69, 0xc6, 0xd8, 0x79, 0x69, 0x12,
			0xae, 0xd6, 0xa4, 0x89, 0xe6, 0xb9, 0xf1, 0xdd,
			0x7a, 0xea, 0x4e, 0x5a, 0x49, 0x7e, 0x6e, 0xe5,
			0xc2, 0xb4, 0x05, 0x05, 0x11, 0xd9, 0xda, 0x9d,
		}
		expectDecrypted := []uint8{
			0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		}
		a := require.New(t)
		decrypted, err := decryptLocal(expectEncrypted, key)
		a.NoError(err)
		a.Equal(expectDecrypted, decrypted)

		encrypted, err := encryptLocal(decrypted, key)
		a.NoError(err)
		a.Equal(expectEncrypted, encrypted)

		decrypted2, err := decryptLocal(encrypted, key)
		a.NoError(err)
		a.Equal(expectDecrypted, decrypted2)
	})
}
