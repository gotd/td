package mtproto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nnqq/td/internal/proto"
)

func TestClientNewMessageID(t *testing.T) {
	c := newTestClient(nil)
	now := c.clock.Now()
	id := proto.MessageID(c.newMessageID())
	assert.Equal(t, proto.MessageFromClient, id.Type())

	lag := id.Time().Sub(now)
	if lag < 0 {
		lag *= -1
	}
	if lag > time.Second {
		t.Error("generated id lags in time")
	}
}
