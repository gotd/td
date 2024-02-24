package gen

import (
	"bytes"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/go-faster/errors"
)

// config is input data for templates.
type config struct {
	Layer      int
	Flags      GenerateFlags
	Package    string
	Structs    []structDef
	Interfaces []interfaceDef
	Mappings   map[string][]constructorMapping
	Registry   []bindingDef
	Errors     []errCheckDef
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

func (g *Generator) shouldGenerateClassifier() bool {
	for _, s := range g.structs {
		if s.Interface == "MessageClass" {
			return g.hasUpdateClass()
		}
	}
	return false
}

func (g *Generator) hasUpdateClass() bool {
	for _, s := range g.structs {
		if s.Interface == "UpdateClass" {
			return true
		}
	}
	return false
}

type writer struct {
	pkg   string
	fs    FileSystem
	t     *template.Template
	buf   *bytes.Buffer
	wrote map[string]bool

	wroteConstructors map[string]struct{}
	generateFlags     GenerateFlags
}

// Generate executes template to file using config.
func (w *writer) Generate(templateName, fileName string, cfg config) error {
	if cfg.Package == "" {
		cfg.Package = w.pkg
	}
	if w.wrote[fileName] {
		return errors.Errorf("name collision (already wrote %s)", fileName)
	}

	w.buf.Reset()
	if err := w.t.ExecuteTemplate(w.buf, templateName, cfg); err != nil {
		return errors.Wrapf(err, "execute template %s for %s", templateName, fileName)
	}
	if err := w.fs.WriteFile(fileName, w.buf.Bytes()); err != nil {
		_ = os.WriteFile(fileName+".dump", w.buf.Bytes(), 0600)
		return errors.Wrapf(err, "write file %s", fileName)
	}
	w.wrote[fileName] = true

	return nil
}

func (w *writer) write(fileName string, cfg config) error {
	if err := w.Generate("main", fileName, cfg); err != nil {
		return err
	}

	if w.generateFlags.Slices {
		name := strings.TrimSuffix(fileName, "_gen.go") + "_slices_gen.go"
		if err := w.Generate("slices", name, cfg); err != nil {
			return err
		}
	}

	return nil
}

// WriteInterfaces writes interface definitions to corresponding files.
func (w *writer) WriteInterfaces(interfaces []interfaceDef) error {
	for _, class := range interfaces {
		cfg := config{
			Package:    w.pkg,
			Structs:    class.Constructors,
			Interfaces: []interfaceDef{class},
			Flags:      w.generateFlags,
		}
		for _, s := range cfg.Structs {
			w.wroteConstructors[s.Name] = struct{}{}
		}

		name := outFileName(class.BaseName, class.Namespace)
		if err := w.write(name, cfg); err != nil {
			return err
		}
	}
	return nil
}

// WriteStructs writes structure definitions to corresponding files.
func (w *writer) WriteStructs(structs []structDef, mappings map[string][]constructorMapping) error {
	for _, s := range structs {
		if _, ok := w.wroteConstructors[s.Name]; ok {
			continue
		}
		cfg := config{
			Package:  w.pkg,
			Structs:  []structDef{s},
			Mappings: mappings,
			Flags:    w.generateFlags,
		}
		name := outFileName(s.BaseName, s.Namespace)
		if w.wrote[name] {
			// Name collision.
			name = outFileName(s.BaseName+"_const", s.Namespace)
		}
		if err := w.write(name, cfg); err != nil {
			return err
		}
	}

	return nil
}

// WriteSource writes generated definitions to fs.
func (g *Generator) WriteSource(fs FileSystem, pkgName string, t *template.Template) error {
	w := &writer{
		pkg:   pkgName,
		fs:    fs,
		t:     t,
		buf:   new(bytes.Buffer),
		wrote: map[string]bool{},

		wroteConstructors: map[string]struct{}{},
		generateFlags:     g.generateFlags,
	}

	if err := w.WriteInterfaces(g.interfaces); err != nil {
		return errors.Wrap(err, "interfaces")
	}
	if err := w.WriteStructs(g.structs, g.mappings); err != nil {
		return errors.Wrap(err, "structs")
	}
	if g.generateFlags.Server {
		if err := w.Generate("server", "tl_server_gen.go", config{
			Structs: g.structs,
		}); err != nil {
			return err
		}
	}

	if g.generateFlags.Handlers && g.hasUpdateClass() {
		if err := w.Generate("handlers", "tl_handlers_gen.go", config{
			Structs: g.structs,
		}); err != nil {
			return err
		}
	}
	if g.generateFlags.UpdatesClassifier && g.shouldGenerateClassifier() {
		if err := w.Generate("updates_classifier", "tl_updates_classifier_gen.go", config{
			Structs: g.structs,
		}); err != nil {
			return err
		}
	}

	sort.SliceStable(g.interfaces, func(i, j int) bool {
		return g.interfaces[i].Name < g.interfaces[j].Name
	})
	cfg := config{
		Registry:   g.registry,
		Interfaces: g.interfaces,
		Layer:      g.schema.Layer,
		Errors:     g.errorChecks,
		Flags:      g.generateFlags,
	}

	if g.generateFlags.Registry {
		if err := w.Generate("registry", "tl_registry_gen.go", cfg); err != nil {
			return err
		}
	}
	if g.generateFlags.Client {
		if err := w.Generate("client", "tl_client_gen.go", cfg); err != nil {
			return err
		}
	}
	if len(cfg.Errors) > 0 {
		if err := w.Generate("errors", "tl_errors_gen.go", cfg); err != nil {
			return err
		}
	}

	return nil
}
