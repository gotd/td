package telegram_test

import (
	"testing"
	"time"

	"github.com/ogen-go/errors"
	"github.com/stretchr/testify/assert"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tgerr"
)

func TestAsFloodWait(t *testing.T) {
	err := func() error {
		return errors.Wrap(tgerr.New(400, "FLOOD_WAIT_10"), "perform operation")
	}()

	d, ok := telegram.AsFloodWait(err)
	assert.True(t, ok)
	assert.Equal(t, time.Second*10, d)
}
