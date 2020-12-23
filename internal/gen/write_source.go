package gen

import (
	"bytes"
	"io"
	"os"
	"strings"
	"text/template"

	"golang.org/x/xerrors"
)

// config is input data for templates.
type config struct {
	Layer      int
	Package    string
	Structs    []structDef
	Interfaces []interfaceDef
	Registry   []bindingDef
}

// FileSystem represents a directory of generated package.
type FileSystem interface {
	WriteFile(baseName string, source []byte) error
}

// outFileName returns file name of generated go source file based on namespace
// and baseName in snake case.
func outFileName(baseName string, namespace []string) string {
	var s strings.Builder
	s.WriteString("tl_")
	for _, ns := range namespace {
		s.WriteString(rules.Underscore(ns))
		s.WriteString("_")
	}
	s.WriteString(rules.Underscore(baseName))
	s.WriteString("_gen.go")
	return s.String()
}

func (g *Generator) hasUpdateClass() bool {
	for _, s := range g.structs {
		if s.Interface == "UpdateClass" {
			return true
		}
	}
	return false
}

// WriteSource writes generated definitions to fs.
func (g *Generator) WriteSource(fs FileSystem, pkgName string, t *template.Template) error {
	buf := new(bytes.Buffer)
	wrote := make(map[string]bool)
	generate := func(templateName, name string, cfg config) error {
		if wrote[name] {
			return xerrors.Errorf("name collision (already wrote %s)", name)
		}

		buf.Reset()
		if err := t.ExecuteTemplate(buf, templateName, cfg); err != nil {
			return xerrors.Errorf("failed to execute template %s for %s: %w", templateName, name, err)
		}
		if err := fs.WriteFile(name, buf.Bytes()); err != nil {
			_, _ = io.Copy(os.Stderr, buf)
			return xerrors.Errorf("failed to write file %s: %w", name, err)
		}
		wrote[name] = true

		return nil
	}

	wroteConstructors := make(map[string]struct{})
	for _, class := range g.interfaces {
		cfg := config{
			Package:    pkgName,
			Interfaces: []interfaceDef{class},
			Structs:    class.Constructors,
		}
		for _, s := range cfg.Structs {
			wroteConstructors[s.Name] = struct{}{}
		}

		name := outFileName(class.BaseName, class.Namespace)
		if err := generate("main", name, cfg); err != nil {
			return err
		}
	}
	for _, s := range g.structs {
		if _, ok := wroteConstructors[s.Name]; ok {
			continue
		}
		cfg := config{
			Package: pkgName,
			Structs: []structDef{s},
		}
		name := outFileName(s.BaseName, s.Namespace)
		if wrote[name] {
			// Name collision.
			name = outFileName(s.BaseName+"_const", s.Namespace)
		}
		if err := generate("main", name, cfg); err != nil {
			return err
		}
	}

	if g.hasUpdateClass() {
		cfg := config{
			Package: pkgName,
			Structs: g.structs,
		}
		if err := generate("handlers", "tl_handlers_gen.go", cfg); err != nil {
			return err
		}
	}

	cfg := config{
		Package:  pkgName,
		Registry: g.registry,
		Layer:    g.schema.Layer,
	}

	if err := generate("registry", "tl_registry_gen.go", cfg); err != nil {
		return err
	}

	if err := generate("client", "tl_client_gen.go", cfg); err != nil {
		return err
	}
	return nil
}
