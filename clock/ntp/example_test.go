package ntp_test

import (
	"fmt"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/clock/ntp"
	"github.com/gotd/td/telegram"
)

func ExampleNew() {
	// Create an NTP-calibrated clock.
	c, err := ntp.New(ntp.Options{
		Server: ntp.DefaultServer,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("clock offset: %s\n", c.Offset())

	// Use it as the time source for the client.
	_ = telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
		Clock: c,
	})

	// Re-calibrate periodically if the process runs for a long time.
	_ = c.Sync()

	var _ clock.Clock = c
}
