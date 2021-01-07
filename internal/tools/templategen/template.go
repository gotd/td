package main

import (
	"fmt"
	"text/template"

	"github.com/gotd/td/internal/gen"
)

func checkTemplates(config config) error {
	for _, dir := range config.Input {
		_, err := template.New("").Funcs(gen.Funcs()).ParseGlob(glob(dir))
		if err != nil {
			return fmt.Errorf("parse %s failed: %w", dir.Path, err)
		}
	}

	return nil
}
