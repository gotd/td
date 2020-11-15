package parser

import (
	"strings"
	"unicode"

	"golang.org/x/xerrors"
)

// Annotation represents an annotation comment, like //@name value.
type Annotation struct {
	// Name of annotation.
	//
	// Can be:
	//	* "description" if Value is class or definition description
	//	* "class" if Value is class name, like //@class Foo @description Foo class
	//	* "param_description" if Value is description for "description" parameter
	//	Otherwise, it is description of Name parameter.
	Name string `json:"name"`
	// Value of annotation. Can be description or class name if Name is "class".
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

// singleLineAnnotations encodes multiple annotations on single line.
//
// NB: newlines are not quited if present.
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

// parseAnnotation parses one or multiple annotations on the line.
func parseAnnotation(line string) ([]Annotation, error) {
	//@name The name of the option @parserValue The new parserValue of the option
	if !strings.HasPrefix(line, "//") {
		return nil, xerrors.New("annotation should be comment")
	}
	line = strings.TrimSpace(strings.TrimLeft(line, "/"))
	if line == "" {
		return nil, xerrors.New("blank comment")
	}
	if !strings.HasPrefix(line, "@") {
		return nil, xerrors.New("invalid annotation start")
	}

	// Probably this can be simplified.
	var annotations []Annotation
	for line != "" {
		nameEnd := strings.Index(line, " ")
		if nameEnd <= 1 {
			return nil, xerrors.New("failed to find name end")
		}
		name := line[1:nameEnd]
		if !isValidName(name) {
			return nil, xerrors.New("invalid annotation name")
		}

		line = line[nameEnd:]
		nextAnnotationPos := strings.Index(line, "@")
		if nextAnnotationPos < 0 {
			// No more annotations.
			value := strings.TrimSpace(line)
			if !isValidAnnotationValue(value) {
				return nil, xerrors.Errorf("invalid annotation value %q", value)
			}
			annotations = append(annotations, Annotation{
				Name:  name,
				Value: value,
			})
			break
		}

		// There will be more.
		value := strings.TrimSpace(line[:nextAnnotationPos])
		if !isValidAnnotationValue(value) {
			return nil, xerrors.Errorf("invalid annotation value %q", value)
		}
		annotations = append(annotations, Annotation{
			Name:  name,
			Value: value,
		})
		line = line[nextAnnotationPos:]
	}
	return annotations, nil
}

func isValidAnnotationValue(v string) bool {
	if v == "" {
		return false
	}
	for _, s := range v {
		if unicode.IsControl(s) {
			return false
		}
		if unicode.IsPrint(s) {
			continue
		}
		return false
	}
	return true
}
