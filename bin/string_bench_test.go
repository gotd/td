package bin_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func BenchmarkSmallString(b *testing.B) {
	corpus := []string{
		"tdakkota", "ernado", "xjem", "russ",
		"botapi", "mtproto", "go", "intern",
		"cox", "gotd", "десять", "ура давай ура",
	}
	buffers := make([]bin.Buffer, len(corpus))
	for i := range buffers {
		if err := buffers[i].Encode(&tg.HelpUserInfo{
			Message: corpus[rand.Intn(len(corpus))],
			Author:  corpus[rand.Intn(len(corpus))],
			Date:    int(time.Now().Unix()),
		}); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf := buffers[rand.Intn(len(buffers))]

		var p tg.HelpUserInfo
		if err := p.Decode(&bin.Buffer{Buf: buf.Buf}); err != nil {
			b.Fatal(err)
		}
	}
}
