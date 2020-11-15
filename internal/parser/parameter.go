package parser

import (
	"fmt"
	"strings"

	"golang.org/x/xerrors"
)

// Parameter with Name and Type.
type Parameter struct {
	// Name of Parameter.
	Name string `json:"name,omitempty"`
	// Type of Parameter.
	Type Type `json:"type"`
	// Flag specifies flag name and index if parameter is conditional.
	Flag *Flag `json:"flag,omitempty"`
	// Flags denotes whether Parameter is flags field (uint32).
	//
	// If true, Type and Flag are blank.
	Flags bool `json:"flags,omitempty"`

	// special case for {X:Type} definitions aka generic parameters,
	// only "Name" field is populated.
	typeDefinition bool
}

func (p Parameter) Conditional() bool {
	return p.Flag != nil
}

func (p Parameter) String() string {
	var b strings.Builder
	if p.typeDefinition {
		b.WriteRune('{')
	}
	if p.Name != "" {
		// Anonymous parameter.
		b.WriteString(p.Name)
		b.WriteRune(':')
	}
	if p.typeDefinition {
		b.WriteString("Type}")
		return b.String()
	}
	if p.Flags {
		b.WriteRune('#')
		return b.String()
	}
	if p.Conditional() {
		b.WriteString(p.Flag.String())
		b.WriteRune('?')
	}
	b.WriteString(p.Type.String())
	return b.String()
}

func (p *Parameter) Parse(s string) error {
	if strings.HasPrefix(s, "{") {
		if !strings.HasSuffix(s, ":Type}") {
			return xerrors.Errorf("unexpected generic %s", s)
		}
		p.typeDefinition = true
		p.Name = strings.SplitN(s[1:], ":", 2)[0]
		return nil
	}
	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 2 {
		p.Name = parts[0]
		s = parts[1]
	} else {
		// Anonymous parameter.
		s = parts[0]
	}

	if s == "#" {
		p.Flags = true
		return nil
	}
	if pos := strings.Index(s, "?"); pos >= 0 {
		p.Flag = &Flag{}
		if err := p.Flag.Parse(s[:pos]); err != nil {
			return xerrors.Errorf("failed to parse flag: %w", err)
		}
		s = s[pos+1:]
	}
	if err := p.Type.Parse(s); err != nil {
		return xerrors.Errorf("failed to parse type: %w", err)
	}
	return nil
}

// Compile-time interface implementation assertion.
var (
	_ fmt.Stringer = Parameter{}
)
