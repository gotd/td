package telegram

import (
	"runtime/debug"
	"strings"
	"sync"
)

var versionOnce struct {
	version string
	sync.Once
}

// getVersion optimistically gets current client version.
//
// Does not handle replace directives.
func getVersion() string {
	versionOnce.Do(func() {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return
		}
		// Hard-coded package name. Probably we can generate this via parsing
		// the go.mod file.
		const pkg = "github.com/gotd/td"
		for _, d := range info.Deps {
			if strings.HasPrefix(d.Path, pkg) {
				versionOnce.version = d.Version
				break
			}
		}
	})

	return versionOnce.version
}
