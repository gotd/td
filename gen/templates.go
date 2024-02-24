package gen

import (
	"embed"
	"strings"
	"text/template"
)

var goKeywords = map[string]struct{}{
	// See https://golang.org/ref/spec#Keywords.
	"break":       {},
	"default":     {},
	"func":        {},
	"interface":   {},
	"select":      {},
	"case":        {},
	"defer":       {},
	"go":          {},
	"map":         {},
	"struct":      {},
	"chan":        {},
	"else":        {},
	"goto":        {},
	"package":     {},
	"switch":      {},
	"const":       {},
	"fallthrough": {},
	"if":          {},
	"range":       {},
	"type":        {},
	"continue":    {},
	"for":         {},
	"import":      {},
	"return":      {},
	"var":         {},

	// Not really keyword, but unlikely to shadow.
	// See go/types/universe.go.
	"append":  {},
	"cap":     {},
	"close":   {},
	"complex": {},
	"copy":    {},
	"delete":  {},
	"imag":    {},
	"len":     {},
	"make":    {},
	"new":     {},
	"panic":   {},
	"print":   {},
	"println": {},
	"real":    {},
	"recover": {},
}

// Funcs returns functions which used in templates.
func Funcs() template.FuncMap {
	return template.FuncMap{
		"trim":                 strings.TrimSpace,
		"lower":                strings.ToLower,
		"trimPrefix":           strings.TrimPrefix,
		"trimSuffix":           strings.TrimSuffix,
		"hasPrefix":            strings.HasPrefix,
		"hasSuffix":            strings.HasSuffix,
		"contains":             strings.Contains,
		"hasField":             hasField,
		"optionalField":        optionalField,
		"mapCollectableFields": mapCollectableFields,
		"sortableFields":       sortableFields,
		"generateSliceHelper":  generateSliceHelper,
		"concat": func(args ...interface{}) []interface{} {
			return args
		},
		"add": func(x, y int) int {
			return x + y
		},
		"notEmpty": func(s string) bool {
			return strings.TrimSpace(s) != ""
		},
		"lowerGo": func(input string) string {
			lower := strings.ToLower(input)
			if _, ok := goKeywords[lower]; ok {
				return lower + "_"
			}
			return lower
		},
		"hasFlags": func(def structDef) bool {
			for _, field := range def.Fields {
				if field.Type == flagsType {
					return true
				}
			}
			return false
		},

		// Argument constructors
		"newStructConfig":    newStructConfig,
		"newInterfaceConfig": newInterfaceConfig,
	}
}

//go:embed _template/*.tmpl
var templates embed.FS

// Template parses and returns vendored code generation templates.
func Template() *template.Template {
	tmpl := template.New("templates").Funcs(Funcs())
	tmpl = template.Must(tmpl.ParseFS(templates, "_template/*.tmpl"))
	return tmpl
}
