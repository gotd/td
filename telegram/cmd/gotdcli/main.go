package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/ernado/td/telegram"
)

func main() {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()
	defer func() { _ = logger.Sync() }()

	client, err := telegram.Dial(ctx, telegram.Options{
		Addr:   "149.154.167.40:443",
		Logger: logger,
	})
	if err != nil {
		panic(err)
	}
	if err := client.Connect(ctx); err != nil {
		panic(err)
	}
	if err := client.CreateAuthKey(ctx); err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
	logger.Info("Created auth key")

	if err := client.Ping(ctx); err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
	logger.Info("Ping ok")

	if err := client.InitConnection(ctx); err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
	logger.Info("Connection initialized")
}
