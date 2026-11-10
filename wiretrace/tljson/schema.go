package tljson

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-faster/errors"

	"github.com/gotd/tl"
)

// Well-known constructor ids that are parsed manually (they are commented
// out or special in the schema).
const (
	IDVector       uint32 = 0x1cb5c415
	IDRPCResult    uint32 = 0xf35c6d01
	IDRPCError     uint32 = 0x2144ca19
	IDMsgContainer uint32 = 0x73f1f8dc
	IDGzipPacked   uint32 = 0x3072cfa1
	IDMsgCopy      uint32 = 0xe06046b2
	IDBoolTrue     uint32 = 0x997275b5
	IDBoolFalse    uint32 = 0xbc799737
)

// Def is a schema definition (constructor or function).
type Def struct {
	// Name is the full name, with namespace: "messages.getDialogs".
	Name     string
	ID       uint32
	Params   []tl.Parameter
	Type     tl.Type
	Function bool
	// Generic is the set of generic parameter names ({X:Type}).
	Generic []string
}

// Schema is a set of TL definitions indexed for decoding.
type Schema struct {
	// Layer is the "// LAYER N" annotation, if present.
	Layer int
	// Files are the sha256 (hex) of the parsed sources, keyed by base name.
	Files map[string]string
	// Skipped counts definition lines the parser could not handle (built-in
	// definitions like `int128 4*[ int ] = Int128`).
	Skipped int

	constructors map[uint32]*Def
	functions    map[uint32]*Def
	byName       map[string]*Def
	funcByName   map[string]*Def
	// byType lists constructors of a boxed type, by full type name.
	byType map[string][]*Def
}

// NewSchema returns an empty schema.
func NewSchema() *Schema {
	return &Schema{
		Files:        map[string]string{},
		constructors: map[uint32]*Def{},
		functions:    map[uint32]*Def{},
		byName:       map[string]*Def{},
		funcByName:   map[string]*Def{},
		byType:       map[string][]*Def{},
	}
}

// LoadFiles parses TL files (e.g. api.tl and mtproto.tl) into one schema.
func LoadFiles(paths ...string) (*Schema, error) {
	s := NewSchema()
	for _, p := range paths {
		f, err := os.Open(p) // #nosec G304
		if err != nil {
			return nil, errors.Wrap(err, "open schema")
		}
		h := sha256.New()
		err = s.Parse(io.TeeReader(f, h))
		_ = f.Close()
		if err != nil {
			return nil, errors.Wrapf(err, "parse %s", p)
		}
		s.Files[filepath.Base(p)] = hex.EncodeToString(h.Sum(nil))
	}
	return s, nil
}

// SchemaFiles are the file names LoadDir looks for.
var SchemaFiles = []string{"api.tl", "mtproto.tl"}

// LoadDir loads api.tl and mtproto.tl from dir.
func LoadDir(dir string) (*Schema, error) {
	var paths []string
	for _, name := range SchemaFiles {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err != nil {
			return nil, errors.Wrapf(err, "schema dir %s", dir)
		}
		paths = append(paths, p)
	}
	return LoadFiles(paths...)
}

func fullName(ns []string, name string) string {
	if len(ns) == 0 {
		return name
	}
	return strings.Join(ns, ".") + "." + name
}

// TypeName returns the full name of a type, without generic arguments.
func TypeName(t tl.Type) string {
	return fullName(t.Namespace, t.Name)
}

// Parse adds definitions from a TL source.
//
// Parsing is line based, like tl.Parse, but lines that tl.Definition can't
// parse are skipped (and counted) instead of failing the whole file: the
// tdesktop mtproto.tl has built-in definitions like `int128 4*[ int ]`.
func (s *Schema) Parse(r io.Reader) error {
	const layerPrefix = "// LAYER "
	function := false
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		switch {
		case line == "":
			continue
		case line == "---functions---":
			function = true
			continue
		case line == "---types---":
			function = false
			continue
		case strings.HasPrefix(line, layerPrefix):
			if v, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, layerPrefix))); err == nil {
				s.Layer = v
			}
			continue
		case strings.HasPrefix(line, "//"):
			continue
		}
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		var d tl.Definition
		if err := d.Parse(line); err != nil || strings.Contains(line, "[") {
			s.Skipped++
			continue
		}
		if d.Name == typeVector || d.Base {
			continue
		}
		s.Add(&Def{
			Name:     fullName(d.Namespace, d.Name),
			ID:       d.ID,
			Params:   d.Params,
			Type:     d.Type,
			Function: function,
			Generic:  d.GenericParams,
		})
	}
	if err := sc.Err(); err != nil {
		return errors.Wrap(err, "scan")
	}
	return nil
}

// Add adds a definition, replacing the previous one with the same id.
func (s *Schema) Add(d *Def) {
	if d.Function {
		s.functions[d.ID] = d
		s.funcByName[d.Name] = d
		return
	}
	s.constructors[d.ID] = d
	s.byName[d.Name] = d
	tn := TypeName(d.Type)
	s.byType[tn] = append(s.byType[tn], d)
}

// Constructor returns a constructor by id.
func (s *Schema) Constructor(id uint32) (*Def, bool) {
	d, ok := s.constructors[id]
	return d, ok
}

// FunctionByName returns a function by full name.
func (s *Schema) FunctionByName(name string) (*Def, bool) {
	d, ok := s.funcByName[name]
	return d, ok
}

// GenericParam returns the name of the `!X` parameter of a generic function
// (the wrapped query of invokeWithLayer, initConnection, ...).
func (d *Def) GenericParam() (string, bool) {
	for _, p := range d.Params {
		if p.Type.GenericRef {
			return p.Name, true
		}
	}
	return "", false
}

// Function returns a function by id.
func (s *Schema) Function(id uint32) (*Def, bool) {
	d, ok := s.functions[id]
	return d, ok
}

// Lookup returns a constructor or a function by id, preferring functions if
// function is true.
func (s *Schema) Lookup(id uint32, function bool) (*Def, bool) {
	if function {
		if d, ok := s.functions[id]; ok {
			return d, true
		}
		d, ok := s.constructors[id]
		return d, ok
	}
	if d, ok := s.constructors[id]; ok {
		return d, true
	}
	d, ok := s.functions[id]
	return d, ok
}

// Name returns the name of a constructor or function id, or "#xxxxxxxx".
func (s *Schema) Name(id uint32) string {
	switch id {
	case IDVector:
		return typeVector
	case IDRPCResult:
		return "rpc_result"
	case IDMsgContainer:
		return "msg_container"
	case IDGzipPacked:
		return "gzip_packed"
	case IDMsgCopy:
		return "msg_copy"
	}
	if s != nil {
		if d, ok := s.Lookup(id, true); ok {
			return d.Name
		}
	}
	return UnknownName(id)
}

// UnknownName formats an unknown constructor id.
func UnknownName(id uint32) string {
	const digits = "0123456789abcdef"
	b := []byte("#00000000")
	for i := 0; i < 8; i++ {
		b[8-i] = digits[(id>>(4*i))&0xf]
	}
	return string(b)
}

// bareConstructor resolves a bare type reference (`future_salt`,
// `%Message`) to its single constructor.
func (s *Schema) bareConstructor(t tl.Type) (*Def, bool) {
	name := TypeName(t)
	if d, ok := s.byName[name]; ok && !t.Percent {
		return d, true
	}
	if list := s.byType[name]; len(list) == 1 {
		return list[0], true
	}
	return nil, false
}
