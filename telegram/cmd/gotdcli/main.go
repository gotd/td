package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/ernado/td/telegram"
)

func readRSAPublicKeys(data []byte) ([]*rsa.PublicKey, error) {
	keys := make([]*rsa.PublicKey, 0)
	for {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}

		key, err := parseRSAFromPEM(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA from PEM: %w", err)
		}

		keys = append(keys, key)
		data = rest
	}

	return keys, nil
}

func readRSAPublicKeysFile(path string) ([]*rsa.PublicKey, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return readRSAPublicKeys(data)
}

func parseRSAFromPEM(data []byte) (*rsa.PublicKey, error) {
	key, err := x509.ParsePKCS1PublicKey(data)
	if err == nil {
		return key, nil
	}
	k, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return nil, err
	}
	kPublic, ok := k.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("parsed unexpected key type %T", k)
	}
	return kPublic, nil
}

func main() {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()
	defer func() { _ = logger.Sync() }()

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	keys, err := readRSAPublicKeysFile(filepath.Join(home, ".td", "public_keys.pem"))
	if err != nil {
		panic(err)
	}
	client, err := telegram.Dial(ctx, telegram.Options{
		Addr:       "149.154.167.40:443",
		PublicKeys: keys,
		Logger:     logger,
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
		panic(err)
	}

	logger.Info("Ping ok")

	if err := client.InitConnection(ctx); err != nil {
		panic(err)
	}

	logger.Info("Connection initialized")
}
