package gen

import (
	"bytes"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/tl"
)

type dirFS string

func (d dirFS) WriteFile(name string, content []byte) error {
	formatted, err := format.Source(content)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(string(d), name), formatted, 0o600)
}

// TestGeneratorHandlersCompile checks that generated update handlers compile
// without hand-written code of the tg package, so --handlers works for any
// target package.
func TestGeneratorHandlersCompile(t *testing.T) {
	if testing.Short() {
		t.Skip("Runs go build")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("No go binary")
	}

	data, err := os.ReadFile(filepath.Join("_testdata", "handlers.tl"))
	require.NoError(t, err)
	schema, err := tl.Parse(bytes.NewReader(data))
	require.NoError(t, err)
	g, err := NewGenerator(schema, GeneratorOptions{
		GenerateFlags: GenerateFlags{
			Handlers: true,
			GetSet:   true,
			Mapping:  true,
			Slices:   true,
		},
	})
	require.NoError(t, err)

	// Inside the module, so imports of github.com/gotd/td resolve. The "_"
	// prefix keeps it out of ./... patterns.
	dir, err := os.MkdirTemp(".", "_handlers_compile")
	if err != nil {
		// E.g. read-only source tree, like the module cache.
		t.Skipf("Can't create package dir in source tree: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	require.NoError(t, g.WriteSource(dirFS(dir), "handlers", Template()))

	cmd := exec.Command(goBin, "vet", "./"+filepath.ToSlash(dir))
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "generated handlers must compile:\n%s", out)
}
