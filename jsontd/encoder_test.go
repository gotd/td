package jsontd_test

import (
	"encoding/json"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/jsontd"
	"github.com/gotd/td/tdapi"
)

func TestEncodeDecode(t *testing.T) {
	a := require.New(t)

	req := &tdapi.SetTdlibParametersRequest{
		Parameters: tdapi.TdlibParameters{
			UseTestDC:              true,
			DatabaseDirectory:      "database",
			FilesDirectory:         "files",
			UseFileDatabase:        true,
			UseChatInfoDatabase:    true,
			UseMessageDatabase:     true,
			UseSecretChats:         true,
			APIID:                  10,
			APIHash:                "russcox",
			SystemLanguageCode:     "ru",
			DeviceModel:            "gotd",
			SystemVersion:          "10",
			ApplicationVersion:     "10",
			EnableStorageOptimizer: true,
			IgnoreFileNames:        true,
		},
	}

	enc := jsontd.Encoder{
		Encoder: jx.GetEncoder(),
	}
	a.NoError(req.EncodeTDLibJSON(enc))
	a.True(json.Valid(enc.Bytes()))

	dec := jsontd.Decoder{
		Decoder: jx.DecodeBytes(enc.Bytes()),
	}
	a.NoError(req.DecodeTDLibJSON(dec))
}
