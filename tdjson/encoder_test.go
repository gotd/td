package tdjson_test

import (
	"encoding/json"
	"math"
	"strconv"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tdapi"
	"github.com/gotd/td/tdjson"
	"github.com/gotd/td/tdp"
)

func TestEncodeDecode(t *testing.T) {
	type obj interface {
		tdjson.TDLibDecoder
		tdjson.TDLibEncoder
		TypeInfo() tdp.Type
	}

	test := func(create func() obj) func(t *testing.T) {
		return func(t *testing.T) {
			a := require.New(t)
			req := create()

			enc := tdjson.Encoder{
				Encoder: jx.GetEncoder(),
			}
			a.NoError(req.EncodeTDLibJSON(enc))
			a.True(json.Valid(enc.Bytes()))

			dec := tdjson.Decoder{
				Decoder: jx.DecodeBytes(enc.Bytes()),
			}
			a.NoError(req.DecodeTDLibJSON(dec))
		}
	}

	types := []obj{
		&tdapi.SetTdlibParametersRequest{
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
		},
		&tdapi.ProfilePhoto{
			ID: 1,
		},
	}

	for _, typ := range types {
		t.Run(typ.TypeInfo().Name, test(func() obj {
			return typ
		}))
	}
}

func TestEncoder_PutLong(t *testing.T) {
	for _, tt := range []int64{
		-1,
		0,
		1,
		10,
		math.MaxInt64,
		math.MinInt64,
	} {
		t.Run(strconv.FormatInt(tt, 10), func(t *testing.T) {
			a := require.New(t)
			e := tdjson.Encoder{Encoder: jx.GetEncoder()}
			e.PutLong(tt)
			data := e.Bytes()
			a.Equal(strconv.Quote(strconv.FormatInt(tt, 10)), string(data))

			d := tdjson.Decoder{Decoder: jx.DecodeBytes(data)}
			v, err := d.Long()
			a.NoError(err)
			a.Equal(tt, v)
		})
	}
}
