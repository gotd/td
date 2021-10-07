package telegram

import (
	"sync"

	"github.com/nnqq/td/internal/mt"
	"github.com/nnqq/td/internal/proto"
	"github.com/nnqq/td/internal/tmap"
	"github.com/nnqq/td/tg"
)

// Port is default port used by telegram.
const Port = 443

var (
	typesMap  *tmap.Map
	typesOnce sync.Once
)

func getTypesMapping() *tmap.Map {
	typesOnce.Do(func() {
		typesMap = tmap.New(
			tg.TypesMap(),
			mt.TypesMap(),
			proto.TypesMap(),
		)
	})
	return typesMap
}
