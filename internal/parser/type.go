package parser

import (
	"encoding"
	"fmt"
	"strings"

	"golang.org/x/xerrors"
)

// Type of a Definition or a Parameter.
type Type struct {
	Namespace  []string // namespace components of the type
	Name       string   // the name of the type
	Bare       bool     // whether this type is bare or boxed
	GenericRef bool     // whether the type name refers to a generic definition
	GenericArg *Type    // generic arguments of the type
}

func (p *Type) UnmarshalText(text []byte) error {
	return p.Parse(string(text))
}

func (p Type) MarshalText() (text []byte, err error) {
	return []byte(p.String()), nil
}

func (p Type) String() string {
	var b strings.Builder
	if p.GenericRef {
		b.WriteRune('!')
	}
	for _, ns := range p.Namespace {
		b.WriteString(ns)
		b.WriteRune('.')
	}
	b.WriteString(p.Name)
	if p.GenericArg != nil {
		b.WriteRune('<')
		b.WriteString(p.GenericArg.String())
		b.WriteRune('>')
	}
	return b.String()
}

func (p *Type) Parse(s string) error {
	if strings.HasPrefix(s, ".") {
		return xerrors.New("type can't start with dot")
	}
	if strings.HasPrefix(s, "!") {
		p.GenericRef = true
		s = s[1:]
	}

	// Parse `type<generic_arg>`
	if pos := strings.Index(s, "<"); pos >= 0 {
		if !strings.HasSuffix(s, ">") {
			return xerrors.New("invalid generic")
		}
		p.GenericArg = &Type{}
		if err := p.GenericArg.Parse(s[pos+1 : len(s)-1]); err != nil {
			return xerrors.Errorf("failed to parse generic: %w", err)
		}
		s = s[:pos]
	}

	// Parse `ns1.ns2.name`
	ns := strings.Split(s, ".")
	if len(ns) == 1 {
		p.Name = ns[0]
	} else {
		p.Name = ns[len(ns)-1]
		p.Namespace = ns[:len(ns)-1]
	}
	if p.Name == "" {
		return xerrors.New("blank name")
	}
	if !isValidName(p.Name) {
		return xerrors.Errorf("invalid name %q", p.Name)
	}
	for _, ns := range p.Namespace {
		if !isValidName(ns) {
			return xerrors.Errorf("invalid namespace part %q", ns)
		}
	}

	// Bare types are always lowercase.
	p.Bare = p.Name == strings.ToLower(p.Name)
	return nil
}

// Compile-time interface implementation assertion.
var (
	_ encoding.TextMarshaler   = Type{}
	_ encoding.TextUnmarshaler = &Type{}
	_ fmt.Stringer             = Type{}
)
