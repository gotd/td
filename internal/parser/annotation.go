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
	line = strings.TrimSpace(strings.TrimLeft(line, "/"))
	if line == "" {
		return nil, xerrors.New("blank comment")
	}
	if !strings.HasPrefix(line, "@") {
		return nil, xerrors.New("invalid annotation start")
	}
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
			annotations = append(annotations, Annotation{
				Name:  name,
				Value: strings.TrimSpace(line),
			})
			break
		}
		// There will be more.
		annotations = append(annotations, Annotation{
			Name:  name,
			Value: strings.TrimSpace(line[:nextAnnotationPos]),
		})
		line = line[nextAnnotationPos:]
	}
	return annotations, nil
}
