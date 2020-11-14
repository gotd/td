package parser

import (
	"encoding"
	"fmt"
	"strings"

	"golang.org/x/xerrors"
)

// Parameter with Name and Type.
type Parameter struct {
	// Name of Parameter.
	Name string
	// Type of Parameter.
	Type Type
	// Flag specifies flag name and index if parameter is conditional.
	Flag Flag
	// Flags denotes whether Parameter is flags field (uint32).
	//
	// If true, Type and Flag are blank.
	Flags bool
}

func (p *Parameter) UnmarshalText(text []byte) error {
	return p.Parse(string(text))
}

func (p Parameter) MarshalText() (text []byte, err error) {
	return []byte(p.String()), nil
}

func (p Parameter) Conditional() bool {
	return p.Flag.Name != ""
}

func (p Parameter) String() string {
	var b strings.Builder
	if p.Name != "" {
		// Anonymous parameter.
		b.WriteString(p.Name)
		b.WriteRune(':')
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
		return xerrors.New("{foo:Type} definitions not supported")
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
	_ encoding.TextMarshaler   = Parameter{}
	_ encoding.TextUnmarshaler = &Parameter{}
	_ fmt.Stringer             = Parameter{}
)
