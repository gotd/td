package mtproto

import (
	"context"
	"testing"

	"github.com/gotd/log"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/tmap"
)

func BenchmarkConn_logWithType(b *testing.B) {
	c := Conn{
		log: log.For(log.Nop),
		types: tmap.New(map[uint32]string{
			0x3fedd339: "true",
		}),
	}
	buf := bin.Buffer{}
	buf.PutID(0x3fedd339)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.logWithType(&buf).Info(context.Background(), "Hi!")
	}
}
