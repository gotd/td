package gen

import (
	"fmt"
	"strings"

	"golang.org/x/xerrors"

	"github.com/gotd/getdoc"
)

// structDef represents go structure definition.
type structDef struct {
	// Name of struct, just like that: `type Name struct {`.
	Name string
	// Comment for struct, in one line.
	Comment string
	// Receiver name. E.g. "m" for Message.
	Receiver string
	// HexID is hex-encoded id, like 1ef134.
	HexID string
	// BufArg is name of Encode and Decode argument of bin.Buffer type
	// that is used in those functions.
	//
	// Should not equal to Name.
	BufArg string
	// RawType is type name from TL schema.
	RawType string

	// Interface refers to interface of generic type.
	Interface     string
	InterfaceFunc string

	// Method name if function definition.
	Method string
	// Result type name.
	Result string
	// ResultSingular denotes whether Result is singular type an can be used
	// directly.
	ResultSingular bool
	// ResultBaseName is BaseName of result interface.
	ResultBaseName string
	ResultFunc     string

	// Fields of structure.
	Fields []fieldDef

	// Namespace for file structure generation.
	Namespace []string
	// BaseName for file structure generation.
	BaseName string

	// URL to documentation.
	// Like https://core.telegram.org/method/account.getPrivacy
	// Or https://core.telegram.org/constructor/account.privacyRules
	URL string

	// Docs is comments from documentation.
	Docs []string
}

type bindingDef struct {
	HexID string // id in hex
	Raw   string // raw tl type
}

func (g *Generator) docStruct(k string) getdoc.Constructor {
	if g.doc == nil {
		return getdoc.Constructor{}
	}
	return g.doc.Constructors[k]
}

func (g *Generator) docMethod(k string) getdoc.Method {
	if g.doc == nil {
		return getdoc.Method{}
	}
	return g.doc.Methods[k]
}

func trimDocs(docs []string) []string {
	var out []string
	for _, s := range docs {
		s = strings.TrimSpace(s)
		out = append(out, strings.Split(s, "\n")...)
	}
	return out
}

// makeStructures generates go structure definition representations.
func (g *Generator) makeStructures() error {
	for _, sd := range g.schema.Definitions {
		var (
			d         = sd.Definition
			typeKey   = definitionType(d)
			docStruct = g.docStruct(typeKey)
			docMethod = g.docMethod(typeKey)
		)
		t, ok := g.types[typeKey]
		if !ok {
			return xerrors.Errorf("failed to find type binding for %q", typeKey)
		}
		if len(sd.Definition.GenericParams) > 0 {
			// TODO(ernado): Support generic params.
			// Such definitions are rare and can be implemented manually.
			continue
		}
		s := structDef{
			Namespace: t.Namespace,
			Name:      t.Name,
			BaseName:  d.Name,

			HexID:   fmt.Sprintf("%x", d.ID),
			BufArg:  "b",
			RawType: fmt.Sprintf("%s#%x", typeKey, d.ID),

			Interface:     t.Interface,
			InterfaceFunc: t.InterfaceFunc,

			Method: t.Method,
			Docs:   docStruct.Description,
		}
		if t.Method != "" {
			s.Docs = docMethod.Description
		}
		s.Docs = trimDocs(s.Docs)
		if g.docBase != nil {
			// Assuming constructor by default.
			s.URL = g.docURL("constructor", typeKey)
		}

		// Selecting receiver based on non-namespaced type.
		s.Receiver = strings.ToLower(d.Name[:1])
		if s.Receiver == "b" {
			// bin.Buffer argument collides with receiver.
			s.BufArg = "buf"
		}
		if s.Comment == "" {
			// TODO(ernado): multi-line comments.
			s.Comment = fmt.Sprintf("%s represents TL type `%s`.", s.Name, s.RawType)
		}
		for _, param := range d.Params {
			f, err := g.makeField(param, sd.Annotations)
			if err != nil {
				return xerrors.Errorf("failed to make field %s: %w", param.Name, err)
			}
			if f.Comment == "" {
				f.Comment = docMethod.Parameters[param.Name]
			}
			if f.Comment == "" {
				f.Comment = docStruct.Fields[param.Name]
			}
			if f.Comment == "" {
				f.Comment = fmt.Sprintf("%s field of %s.", f.Name, s.Name)
			}
			s.Fields = append(s.Fields, f)
		}

		if s.Method != "" && t.Class != "Ok" {
			// RPC call.
			class, ok := g.classes[t.Class]
			if ok {
				s.Result = class.Name
				s.ResultSingular = class.Singular
				s.ResultBaseName = class.BaseName
				s.ResultFunc = class.Func
				s.URL = g.docURL("method", typeKey)
			} else {
				// Not implemented.
				s.Method = ""
			}
		}

		g.structs = append(g.structs, s)
		g.registry = append(g.registry, bindingDef{
			HexID: s.HexID,
			Raw:   s.RawType,
		})
	}

	return nil
}
