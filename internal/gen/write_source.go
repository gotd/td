package gen

import (
	"bytes"
	"strings"
	"text/template"
)

// config is input data for templates.
type config struct {
	RPC bool

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

// WriteSource writes generated definitions to fs.
func (g *Generator) WriteSource(fs FileSystem, pkgName string, t *template.Template) error {
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
		buf := new(bytes.Buffer)
		if err := t.ExecuteTemplate(buf, "main", cfg); err != nil {
			return err
		}
		if err := fs.WriteFile(name, buf.Bytes()); err != nil {
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
		buf := new(bytes.Buffer)
		if err := t.ExecuteTemplate(buf, "main", cfg); err != nil {
			return err
		}
		if err := fs.WriteFile(name, buf.Bytes()); err != nil {
			return err
		}
	}

	cfg := config{
		Package:  pkgName,
		Registry: g.registry,
	}
	buf := new(bytes.Buffer)
	if err := t.ExecuteTemplate(buf, "registry", cfg); err != nil {
		return err
	}
	if err := fs.WriteFile("tl_registry_gen.go", buf.Bytes()); err != nil {
		return err
	}

	buf.Reset()
	if err := t.ExecuteTemplate(buf, "client", cfg); err != nil {
		return err
	}
	if err := fs.WriteFile("tl_client.go", buf.Bytes()); err != nil {
		return err
	}

	return nil
}
