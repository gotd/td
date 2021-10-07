package telegram

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nnqq/td/tg"
)

func Test_parseCDNKeys(t *testing.T) {
	keys := []string{
		`-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA+Lf3PvgE1yxbJUCMaEAk
V0QySTVpnaDjiednB5RbtNWjCeqSVakYHbqqGMIIv5WCGdFdrqOfMNcNSstPtSU6
R9UmRw6tquOIykpSuUOje9H+4XVIKqujyL2ISdK+4ZOMl4hCMkqauw4bP1Sbr03v
ZRQbU6qEA04V4j879BAyBVhr3WG9+Zi+t5XfGSTgSExPYEl8rZNHYNV5RB+BuroV
H2HLTOpT/mJVfikYpgjfWF5ldezV4Wo9LSH0cZGSFIaeJl8d0A8Eiy5B9gtBO8mL
+XfQRKOOmr7a4BM4Ro2de5rr2i2od7hYXd3DO9FRSl4y1zA8Am48Rfd95WHF3N/O
mQIDAQAB
-----END RSA PUBLIC KEY-----`,
		`-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAyu5PXyfp+VFLc2hKJsq/cvQ+wq9V2s1iGMMwcrkXrKAqX0S5QEcY
W9b6pV5LulbsvNcxp/YniiSL4FsAja28B9fH//Y+AolWASomCB0NSVHwS1Pqfe3m
GdLTwDmqU17tSWk/48+Kfn4B+WT85ZIKt8bOnABwnM1AtykX0zKwzm9yKcTX0MeY
rwzgiOQax6J1cfgtLdxl8HVKT6wCOS1e43zpXMU+UoWqRqIan+J6q+ubi1yF4PWl
DyDgJSw8uxlhNNMP4tAnshIRZ1ZZ25O/g58jw1qz5XMztZwLNA2pUxaFtyy1LdHC
FRX7DdwIA/FdOzfWyXYLlCFaSX8K/6CnSQIDAQAB
-----END RSA PUBLIC KEY-----`,
	}

	cdnKeys := make([]tg.CDNPublicKey, 0, len(keys))
	for i, key := range keys {
		cdnKeys = append(cdnKeys, tg.CDNPublicKey{
			DCID:      i + 1,
			PublicKey: key,
		})
	}

	publicKeys, err := parseCDNKeys(cdnKeys...)
	require.NoError(t, err)
	require.Len(t, publicKeys, 2)
}
