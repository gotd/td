//go:build go1.18

package deeplink

import (
	"testing"
)

func addSuites(f *testing.F, suites map[string][]testCase) {
	for _, suite := range suites {
		for _, test := range suite {
			f.Add(test.input)
		}
	}
}

func FuzzParse(f *testing.F) {
	for _, typeSuite := range typeSuites {
		addSuites(f, typeSuite)
	}

	f.Fuzz(func(t *testing.T, link string) {
		_, err := Parse(link)
		if err != nil {
			t.Skip(err)
		}
	})
}
