package parser

import (
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/xerrors"
)

// Definition represents "Type Language" definition.
//
// See https://core.telegram.org/mtproto/TL for reference.
type Definition struct {
	Namespace     []string    `json:"namespace,omitempty"`      // blank if global
	Name          string      `json:"name"`                     // name of definition, aka "predicate" or "method"
	ID            uint32      `json:"id"`                       // crc32(definition) or explicitly specified
	Params        []Parameter `json:"params,omitempty"`         // can be empty
	Type          Type        `json:"type"`                     // type of definition
	Base          bool        `json:"base,omitempty"`           // base type?
	GenericParams []string    `json:"generic_params,omitempty"` // like {T:Type}
}

func (d Definition) String() string {
	var b strings.Builder
	for _, ns := range d.Namespace {
		b.WriteString(ns)
		b.WriteRune('.')
	}
	b.WriteString(fmt.Sprintf("%s#%x", d.Name, d.ID))
	for _, param := range d.GenericParams {
		b.WriteString(" {")
		b.WriteString(param)
		b.WriteString(":Type}")
	}
	for _, param := range d.Params {
		b.WriteRune(' ')
		b.WriteString(param.String())
	}
	if d.Base {
		b.WriteString(" ?")
	}
	b.WriteString(" = ")
	b.WriteString(d.Type.String())
	return b.String()
}

// Parse TL definition line like `foo#123 code:int name:string = Message;`.
func (d *Definition) Parse(line string) error {
	line = strings.TrimRight(line, ";")
	parts := strings.Split(line, "=")
	if len(parts) != 2 {
		return xerrors.New("unexpected definition elements")
	}
	// Splitting definition line into left and right parts.
	// Example: `foo#123 code:int name:string = Message`
	var (
		left  = strings.TrimSpace(parts[0]) // `foo#123 code:int name:string`
		right = strings.TrimSpace(parts[1]) // `Message`
		// Divided left part elements, like []{"foo#123", "code:int", "name:string"}
		leftParts = strings.Split(left, " ")
	)
	if left == "" || right == "" {
		return xerrors.New("definition part is blank")
	}
	if err := d.Type.Parse(right); err != nil {
		return xerrors.Errorf("failed to parse type: %w", err)
	}
	{
		// Parsing definition name and id.
		first := leftParts[0]
		nameParts := strings.SplitN(first, tokID, 2)
		d.Name = nameParts[0]
		if d.Name == "" {
			return xerrors.New("blank name")
		}
		if len(nameParts) > 1 {
			// Parsing definition id as hex to uint32.
			idHex := nameParts[1]
			id, err := strconv.ParseUint(idHex, 16, 32)
			if err != nil {
				return xerrors.Errorf("%s is invalid id: %w", idHex, err)
			}
			d.ID = uint32(id)
		} else {
			// Automatically computing.
			d.ID = crc32.ChecksumIEEE([]byte(line))
		}
		if nsParts := strings.Split(d.Name, "."); len(nsParts) > 1 {
			// Handling definition namespace.
			d.Name = nsParts[len(nsParts)-1]
			d.Namespace = nsParts[:len(nsParts)-1]
		}
		for _, ns := range d.Namespace {
			if !isValidName(ns) {
				return xerrors.Errorf("invalid namespace part %q", ns)
			}
		}
	}
	genericParams := map[string]struct{}{}
	for _, f := range leftParts[1:] {
		// Parsing parameters.
		if f == "?" {
			// Special case.
			d.Base = true
			continue
		}
		var param Parameter
		if err := param.Parse(f); err != nil {
			return xerrors.Errorf("failed to parse param: %w", err)
		}
		// Handling generics.
		// Example:
		// `t#1 {X:Type} x:!X = X;` is valid type definition with generic type "X".
		// Type of parameter "x" is {Name: "X", GenericRef: true}.
		if param.typeDefinition {
			// Parameter is generic type definition like {T:Type}.
			genericParams[param.Name] = struct{}{}
			d.GenericParams = append(d.GenericParams, param.Name)
			continue // not adding generic to actual params
		}
		// Checking that type of generic parameter was defined.
		// E.g. `t#1 {Y:Type} x:!X = X;` is invalid, because X was not defined.
		if param.Type.GenericRef {
			if _, ok := genericParams[param.Type.Name]; !ok {
				return xerrors.Errorf("undefined generic parameter type %s", param.Type.Name)
			}
		}
		d.Params = append(d.Params, param)
	}
	if !isValidName(d.Name) {
		return xerrors.Errorf("invalid name %q", d.Name)
	}
	return nil
}

func isValidName(name string) bool {
	if name == "" {
		return false
	}
	for _, s := range name {
		if unicode.IsDigit(s) {
			continue
		}
		if unicode.IsLetter(s) {
			continue
		}
		return false
	}
	return true
}
