package parser

import (
	"strings"

	"golang.org/x/xerrors"
)

// Annotation represents an annotation comment, like //@name value.
type Annotation struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (a Annotation) String() string {
	var b strings.Builder
	b.WriteString("//")
	b.WriteRune('@')
	b.WriteString(a.Name)
	b.WriteRune(' ')
	b.WriteString(a.Value)
	return b.String()
}

func singleLineAnnotations(a []Annotation) string {
	var b strings.Builder
	for i, ann := range a {
		str := ann.String()
		if i > 0 {
			str = strings.Replace(str, "//", " ", 1)
		}
		b.WriteString(str)
	}
	return b.String()
}

func parseAnnotation(line string) ([]Annotation, error) {
	//@name The name of the option @parserValue The new parserValue of the option
	if !strings.HasPrefix(line, "//") {
		return nil, xerrors.New("annotation should be comment")
	}
	line = strings.TrimLeft(line, "/")
	if line == "" {
		return nil, xerrors.New("blank comment")
	}
	var annotations []Annotation
	for _, p := range strings.Split(line, "@") {
		if p == "" {
			continue
		}
		parts := strings.SplitN(p, " ", 2)
		if len(parts) < 2 {
			continue
		}
		a := Annotation{
			Name:  strings.TrimSpace(parts[0]),
			Value: strings.TrimSpace(parts[1]),
		}
		if !isValidName(a.Name) {
			return annotations, xerrors.Errorf("annotation name %q is invalid", a.Name)
		}
		annotations = append(annotations, a)
	}
	return annotations, nil
}
