package mtproto

import (
	"fmt"

	"github.com/gotd/log"

	"github.com/gotd/td/bin"
)

type logType struct {
	ID   uint32
	Name string
}

// LogAttr returns the type info as an inline log group.
func (l logType) LogAttr() log.Attr {
	attrs := []log.Attr{log.String("type_id", fmt.Sprintf("0x%x", l.ID))}
	if l.Name != "" {
		attrs = append(attrs, log.String("type_name", l.Name))
	}
	return log.Group("", attrs...)
}

func (c *Conn) logWithBuffer(b *bin.Buffer) log.Helper {
	return c.logWithType(b).With(log.Int("size_bytes", b.Len()))
}

func (c *Conn) logWithType(b *bin.Buffer) log.Helper {
	id, err := b.PeekID()
	if err != nil {
		// Type info not available.
		return c.log
	}

	return c.logWithTypeID(id)
}

func (c *Conn) logWithTypeID(id uint32) log.Helper {
	return c.log.With(logType{
		ID:   id,
		Name: c.types.Get(id),
	}.LogAttr())
}
