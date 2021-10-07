// The rsagen command generates rsa.PublicKey variables from PEM-encoded
// RSA public keys.
package main

import (
	"bufio"
	"bytes"
	"crypto/rsa"
	"embed"
	"flag"
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"log"
	"os"
	"text/template"

	"golang.org/x/xerrors"

	"github.com/nnqq/td/internal/crypto"
)

//go:embed _template/*.tmpl
var embedFS embed.FS

var funcs = template.FuncMap{
	"fingerprint": func(pubkey *rsa.PublicKey) string {
		return fmt.Sprintf("%08x", uint64(crypto.RSAFingerprint(pubkey)))
	},
	"chunks": func(b []byte, size int) [][]byte {
		var chunks [][]byte
		for len(b) > size {
			chunks = append(chunks, b[:size])
			b = b[size:]
		}
		if len(b) != 0 {
			chunks = append(chunks, b)
		}
		return chunks
	},
	"single": func(keys []*rsa.PublicKey) (*rsa.PublicKey, error) {
		if count := len(keys); count != 1 {
			return nil, xerrors.Errorf("expected single key, got %d keys", count)
		}
		return keys[0], nil
	},
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	log.SetFlags(log.Llongfile)
	var (
		inPath       = flag.String("f", "", "input path (defaults to stdin)")
		outPath      = flag.String("o", "", "output path (defaults to stdout)")
		tplPath      = flag.String("templates", "", "templates directory (defaults to embed)")
		tplName      = flag.String("exec", "main.tmpl", "template name")
		pkgName      = flag.String("pkg", "main", "package name")
		varName      = flag.String("var", "PK", "variable name")
		singleMode   = flag.Bool("single", false, "emit single key instead of slice")
		formatOutput = flag.Bool("format", true, "run gofmt on output")
		testFunc     = flag.String("test", "", "test function name")
	)
	flag.Parse()

	if *testFunc == "" {
		*testFunc = "TestFingerprint" + *varName
	}

	var err error

	var in []byte
	if *inPath != "" {
		in, err = os.ReadFile(*inPath)
	} else {
		in, err = io.ReadAll(os.Stdin)
	}
	if err != nil {
		log.Printf("read input: %v", err)
		return err
	}

	keys, err := crypto.ParseRSAPublicKeys(in)
	if err != nil {
		log.Printf("parse public keys: %v", err)
		return err
	}

	fsys, _ := fs.Sub(embedFS, "_template")
	if *tplPath != "" {
		fsys = os.DirFS(*tplPath)
	}

	tpl, err := template.New("").Funcs(funcs).ParseFS(fsys, "*.tmpl")
	if err != nil {
		log.Printf("parse templates: %v", err)
		return err
	}

	var inLines [][]byte
	for sc := bufio.NewScanner(bytes.NewReader(in)); sc.Scan(); {
		line := sc.Bytes()
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		inLines = append(inLines, line)
	}

	buf := bytes.NewBuffer(nil)
	if err := tpl.ExecuteTemplate(buf, *tplName, map[string]interface{}{
		"Package":    *pkgName,
		"Variable":   *varName,
		"Keys":       keys,
		"Single":     *singleMode,
		"InputLines": inLines,
		"TestFunc":   *testFunc,
	}); err != nil {
		log.Printf("execute template: %v", err)
		return err
	}

	if *formatOutput {
		p, err := format.Source(buf.Bytes())
		if err != nil {
			log.Printf("format output: %v", err)
			return err
		}
		buf = bytes.NewBuffer(p)
	}

	if *outPath != "" {
		err = os.WriteFile(*outPath, buf.Bytes(), 0o600)
	} else {
		_, err = buf.WriteTo(os.Stdout)
	}
	if err != nil {
		log.Printf("write output: %v", err)
		return err
	}
	return nil
}
