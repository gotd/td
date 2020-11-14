package parser

import (
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/xerrors"
)

type Field struct {
	Name string
	Type string
}

// Definition represents "Type Language" definition.
//
// See https://core.telegram.org/mtproto/TL for reference.
type Definition struct {
	Namespace []string    // blank if global
	Name      string      // name of definition, aka "predicate" or "method"
	ID        uint32      // crc32(definition) or explicitly specified
	Params    []Parameter // can be empty
	Type      Type        // type of definition
	Category  Category    // category of definition (function or type)
	Base      bool        // base type?
}

func (d Definition) String() string {
	var b strings.Builder
	for _, ns := range d.Namespace {
		b.WriteString(ns)
		b.WriteRune('.')
	}
	b.WriteString(fmt.Sprintf("%s#%x", d.Name, d.ID))
	for _, param := range d.Params {
		b.WriteRune(' ')
		b.WriteString(param.String())
	}
	if d.Base {
		b.WriteString(" ?")
	}
	b.WriteString(" =")
	b.WriteString(d.Type.String())
	return b.String()
}

func (d *Definition) Parse(line string) error {
	line = strings.TrimRight(line, ";")
	parts := strings.Split(line, "=")
	if len(parts) != 2 {
		return xerrors.New("unexpected definition elements")
	}
	var (
		left      = strings.TrimSpace(parts[0])
		right     = strings.TrimSpace(parts[1])
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
			d.Name = nsParts[len(nsParts)-1]
			d.Namespace = nsParts[:len(nsParts)-1]
		}
		for _, ns := range d.Namespace {
			if !isValidName(ns) {
				return xerrors.Errorf("invalid namespace part %q", ns)
			}
		}
	}
	for _, f := range leftParts[1:] {
		// Parsing fields.
		if f == "?" {
			// Special case.
			d.Base = true
			continue
		}
		var param Parameter
		if err := param.Parse(f); err != nil {
			return xerrors.Errorf("failed to parse param: %w", err)
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
