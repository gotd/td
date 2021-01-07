package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	c, err := parseFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tasks := []struct {
		name string
		f    func(config) error
	}{
		{"Check templates", checkTemplates},
		{"Create asset", createAsset},
	}

	for _, task := range tasks {
		start := time.Now()
		if err := task.f(c); err != nil {
			fmt.Printf("❌ %q: %s\n", task.name, err)
			return
		}
		fmt.Printf("✓ %q (%s)\n", task.name, time.Since(start))
	}
}
