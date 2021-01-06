package mtproto

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/gotd/td/bin"
)

func (c *Conn) logWithType(b *bin.Buffer) *zap.Logger {
	id, err := b.PeekID()
	if err != nil {
		// Type info not available.
		return c.log
	}

	// Adding hex id of type.
	typeIDStr := fmt.Sprintf("0x%x", id)
	log := c.log.With(zap.String("type_id", typeIDStr))

	// Adding verbose type name if available.
	typeName := c.types.Get(id)
	if typeName != "" {
		log = log.With(zap.String("type_name", typeName))
	}

	return log
}
