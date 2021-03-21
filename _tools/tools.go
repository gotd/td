// +build tools

package tools

import (
	_ "golang.org/x/tools/cmd/stringer"

	_ "github.com/dvyukov/go-fuzz/go-fuzz"
	_ "github.com/dvyukov/go-fuzz/go-fuzz-build"
)
