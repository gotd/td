package main

import "github.com/go-bindata/go-bindata"

func createAsset(config config) error {
	cfg := bindata.NewConfig()
	cfg.Package = config.Package
	cfg.Output = config.Output
	cfg.Input = config.Input
	cfg.ModTime = 1
	cfg.Mode = 420

	return bindata.Translate(cfg)
}
