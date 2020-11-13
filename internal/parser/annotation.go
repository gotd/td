package parser

import (
	"strings"
)

type Annotation struct {
	Key   string
	Value string
}

func parseAnnotation(line string) ([]Annotation, error) {
	//@name The name of the option @value The new value of the option
	line = strings.TrimLeft(line, "/")
	var annotations []Annotation
	for _, p := range strings.Split(line, "@") {
		if p == "" {
			continue
		}
		parts := strings.SplitN(p, " ", 2)
		if len(parts) < 2 {
			continue
		}
		annotations = append(annotations, Annotation{
			Key:   strings.TrimSpace(parts[0]),
			Value: strings.TrimSpace(parts[1]),
		})
	}
	return annotations, nil
}
