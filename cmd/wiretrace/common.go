package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-faster/errors"

	"github.com/gotd/td/mt"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tmap"
	"github.com/gotd/td/wiretrace"
	"github.com/gotd/td/wiretrace/tljson"
)

// schemaEnv overrides the schema directory search.
const schemaEnv = "WIRETRACE_SCHEMA"

// findSchemaDir returns the directory with api.tl and mtproto.tl (or a
// comma-separated list of .tl files): the flag value, $WIRETRACE_SCHEMA, or the run directory and its parents (also
// their "schema" subdirectories).
func findSchemaDir(flagValue, runDir string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if v := os.Getenv(schemaEnv); v != "" {
		return v, nil
	}
	dir, err := filepath.Abs(runDir)
	if err != nil {
		return "", errors.Wrap(err, "abs")
	}
	for i := 0; i < 3; i++ {
		for _, c := range []string{dir, filepath.Join(dir, "schema")} {
			if hasSchema(c) {
				return c, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", errors.Errorf("api.tl and mtproto.tl not found next to %s: pass --schema DIR (or set $%s), "+
		"e.g. tdesktop's Telegram/SourceFiles/mtproto/scheme", runDir, schemaEnv)
}

func isFile(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func hasSchema(dir string) bool {
	for _, name := range tljson.SchemaFiles {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			return false
		}
	}
	return true
}

// loadSchema loads the schema and warns (to stderr) if it does not match the
// trace header.
func loadSchema(stderr io.Writer, flagValue string, run *wiretrace.Run, required bool) (*tljson.Schema, error) {
	dir, err := findSchemaDir(flagValue, run.Dir)
	if err != nil {
		if required {
			return nil, err
		}
		return nil, nil
	}
	var s *tljson.Schema
	if strings.Contains(dir, ",") || isFile(dir) {
		s, err = tljson.LoadFiles(strings.Split(dir, ",")...)
	} else {
		s, err = tljson.LoadDir(dir)
	}
	if err != nil {
		return nil, err
	}
	if h := run.Header; h != nil {
		if h.Layer != 0 && s.Layer != 0 && h.Layer != s.Layer {
			_, _ = fmt.Fprintf(stderr, "wiretrace: warning: trace layer %d, schema layer %d\n", h.Layer, s.Layer)
		}
		for name, want := range map[string]string{"api.tl": h.APISHA256, "mtproto.tl": h.MTProtoSHA256} {
			if want != "" && s.Files[name] != want {
				_, _ = fmt.Fprintf(stderr, "wiretrace: warning: %s sha256 differs from the trace header\n", name)
			}
		}
	}
	return s, nil
}

// namer resolves constructor ids to names: the schema first, then the gotd
// tg/mt maps (layers may differ slightly), then "#xxxxxxxx".
type namer struct {
	schema *tljson.Schema
	gotd   *tmap.Map
}

func newNamer(s *tljson.Schema) *namer {
	return &namer{schema: s, gotd: tmap.New(tg.TypesMap(), mt.TypesMap())}
}

func (n *namer) Name(id uint32) string {
	if id == 0 {
		return ""
	}
	if n.schema != nil {
		if name := n.schema.Name(id); !strings.HasPrefix(name, "#") {
			return name
		}
	}
	switch id {
	case tljson.IDVector:
		return "vector"
	case tljson.IDRPCResult:
		return "rpc_result"
	case tljson.IDMsgContainer:
		return "msg_container"
	case tljson.IDGzipPacked:
		return "gzip_packed"
	}
	if name := n.gotd.Get(id); name != "" {
		if i := strings.IndexByte(name, '#'); i > 0 {
			return name[:i]
		}
		return name
	}
	return tljson.UnknownName(id)
}

func openRun(dir string) (*wiretrace.Run, error) {
	run, err := wiretrace.Open(dir)
	if err != nil {
		return nil, err
	}
	return run, nil
}

// badLines collects unparsable lines.
type badLines struct {
	count int
	first string
}

func (b *badLines) add(segment, line int, err error) {
	b.count++
	if b.first == "" {
		b.first = fmt.Sprintf("segment %d line %d: %v", segment, line, err)
	}
}

func (b *badLines) report(stderr io.Writer) {
	if b.count > 0 {
		_, _ = fmt.Fprintf(stderr, "wiretrace: %d unparsable lines skipped (first: %s)\n", b.count, b.first)
	}
}
