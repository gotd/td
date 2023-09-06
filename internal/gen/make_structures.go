package gen

import (
	"fmt"
	"strings"

	"github.com/go-faster/errors"

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
	// HexID is hex-encoded id, like 29bacabb.
	HexID string
	// BufArg is name of Encode and Decode argument of bin.Buffer type
	// that is used in those functions.
	//
	// Should not equal to Name.
	BufArg string
	// RawType is type name from TL schema, like authPassword#29bacabb.
	RawType string
	// RawName is type name from TL schema without HexID (CRC code), like authPassword.
	RawName string

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
	ResultVector   bool

	UnpackParameters bool
	Vector           bool
	// Fields of structure.
	Fields []fieldDef
	// Tuples of Conditional fields.
	ConditionalTuples [32][]fieldDef

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
	// Links from documentation
	Links []string
	// BotCanUse denotes whether method can be used by bots.
	BotCanUse bool
	// Errors is list of possible errors.
	Errors []getdoc.Error
}

func (s *structDef) fillFromClass(class classBinding) {
	s.Result = class.Name
	s.ResultSingular = class.Singular
	s.ResultBaseName = class.BaseName
	s.ResultFunc = class.Func
	s.ResultVector = class.Vector
}

type bindingDef struct {
	HexID string // id in hex
	Raw   string // raw tl type
	Name  string // go type
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

// makeStructures generates go structure definition representations.
//
// nolint:gocognit,gocyclo // TODO(ernado): simplify
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
			return errors.Errorf("find type binding for %q", typeKey)
		}
		s := structDef{
			Namespace: t.Namespace,
			Name:      t.Name,
			BaseName:  d.Name,

			HexID:   fmt.Sprintf("%x", d.ID),
			BufArg:  "b",
			RawType: fmt.Sprintf("%s#%x", typeKey, d.ID),
			RawName: typeKey,

			Interface:     t.Interface,
			InterfaceFunc: t.InterfaceFunc,

			Method: t.Method,
			Docs:   docStruct.Description,
			Links:  docStruct.Links,
		}
		if t.Method != "" {
			s.Docs = docMethod.Description
			s.Links = docMethod.Links
			s.BotCanUse = docMethod.BotCanUse
			for _, e := range docMethod.Errors {
				if strings.Contains(e.Type, ",") {
					// HACK(ernado): fix in getdoc
					continue
				}
				s.Errors = append(s.Errors, e)
			}
		}
		if g.docBase != nil {
			// Assuming constructor by default.
			s.URL = g.docURL("constructor", typeKey)
		}
		s.Docs = splitLines(s.Docs, g.docLineLimit)

		// Selecting receiver based on non-namespaced type.
		s.Receiver = strings.ToLower(d.Name[:1])
		if s.Receiver == "b" {
			// bin.Buffer argument collides with receiver.
			s.BufArg = "buf"
		}
		if strings.TrimSpace(s.Comment) == "" {
			// TODO(ernado): multi-line comments.
			s.Comment = fmt.Sprintf("%s represents TL type `%s`.", s.Name, s.RawType)
		}

		allFieldRequired := true
		for _, param := range d.Params {
			f, err := g.makeField(param, sd.Annotations)
			if err != nil {
				return errors.Wrapf(err, "make field %s", param.Name)
			}

			f.Links = docStruct.Fields[param.Name].Links
			if t.Method != "" {
				f.Links = docMethod.Parameters[param.Name].Links
			}

			if f.Conditional {
				if f.ConditionalIndex >= 32 || f.ConditionalIndex < 0 {
					return errors.Errorf("invalid conditional index %d", f.ConditionalIndex)
				}
				s.ConditionalTuples[f.ConditionalIndex] = append(s.ConditionalTuples[f.ConditionalIndex], f)
				allFieldRequired = false
			}

			if len(f.Comment) < 1 || f.Comment[0] == "" {
				comment := docMethod.Parameters[param.Name].Description
				if strings.TrimSpace(comment) == "" {
					comment = docStruct.Fields[param.Name].Description
				}
				if strings.TrimSpace(comment) == "" {
					comment = fmt.Sprintf("%s field of %s.", f.Name, s.Name)
				}

				f.Comment = []string{comment}
			}

			f.Comment = splitLines(f.Comment, g.docLineLimit)
			s.Fields = append(s.Fields, f)
		}

		if allFieldRequired && len(d.Params) < 2 {
			s.UnpackParameters = true
		}

		if s.Method != "" && t.Class != "Ok" {
			// RPC call.
			class, ok := g.classes[t.Class]

			switch {
			case !ok && strings.HasPrefix(t.Class, "Vector<"):
				var err error
				class, err = g.makeVector(t.Class)
				if err != nil {
					return err
				}
				fallthrough
			case ok:
				s.fillFromClass(class)
				s.URL = g.docURL("method", typeKey)
			default:
				// Not implemented.
				s.Method = ""
			}
		}

		g.structs = append(g.structs, s)
		g.registry = append(g.registry, bindingDef{
			HexID: s.HexID,
			Raw:   s.RawType,
			Name:  s.Name,
		})
	}

	return nil
}
