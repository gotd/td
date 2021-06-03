package session_test

import (
	"context"
	"fmt"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
)

func ExampleTelethonSession() {
	// Get a session from Telethon.
	str := `1AsCoAAEBu2FhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYW
FhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhY
WFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFh
YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE=`

	data, err := session.TelethonSession(str)
	if err != nil {
		panic(err)
	}

	fmt.Println(data.DC, data.Addr)
	// Output:
	// 2 192.168.0.1:443
}

func ExampleTelethonSession_convert() {
	ctx := context.Background()
	// Get a session from Telethon.
	str := `1AsCoAAEBu2FhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYW
FhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhY
WFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFh
YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE=`

	data, err := session.TelethonSession(str)
	if err != nil {
		panic(err)
	}

	var (
		storage = new(session.StorageMemory)
		loader  = session.Loader{Storage: storage}
	)

	// Save decoded Telethon session as gotd session.
	if err := loader.Save(ctx, data); err != nil {
		panic(err)
	}

	// Create client.
	telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
		SessionStorage: storage,
	})

	// Load saved gotd session.
	saved, err := loader.Load(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(saved.DC, saved.Addr)
	// Output:
	// 2 192.168.0.1:443
}
