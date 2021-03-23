package keyparser

import (
	"bytes"
	"embed"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed _testdata
var testData embed.FS

func read(t testing.TB, name string) string {
	t.Helper()

	buf, err := testData.ReadFile(path.Join("_testdata", name))
	if err != nil {
		t.Fatal(err)
	}

	return string(buf)
}

func TestExtract(t *testing.T) {
	var (
		out = new(bytes.Buffer)
		in  = strings.NewReader(read(t, "input.cpp"))
	)
	if err := Extract(in, out); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, read(t, "output.pem"), out.String())
}
