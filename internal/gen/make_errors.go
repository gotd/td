package gen

import (
	"sort"
	"strings"

	"github.com/gotd/getdoc"
)

// errCheckDef is helper for checking errors.
type errCheckDef struct {
	// Name of check without "Is" prefix.
	Name string
	// Type of error.
	Type string
}

func (Generator) errDef(err getdoc.Error) errCheckDef {
	var parts []string
	for _, p := range strings.Split(err.Type, "_") {
		switch p {
		case "X", "*", "%d":
			continue
		default:
			p = strings.ReplaceAll(p, "%d", "")
			parts = append(parts, p)
		}
	}
	var partsLower []string
	for _, p := range parts {
		partsLower = append(partsLower, strings.ToLower(p))
	}
	return errCheckDef{
		Name: pascalWords(partsLower),
		Type: strings.Join(parts, "_"),
	}
}

// makeErrors created go definitions for possible errors.
func (g *Generator) makeErrors() {
	if g.doc == nil {
		return
	}

	// For each unique error Type, create error check definition.
	// Like IsNeedMigration(err) function.
	seen := make(map[string]struct{})
	for _, m := range g.structs {
		for _, e := range m.Errors {
			d := g.errDef(e)

			if _, ok := seen[d.Type]; ok {
				continue
			}
			seen[d.Type] = struct{}{}

			g.errorChecks = append(g.errorChecks, d)
		}
	}

	// Ensure error order.
	sort.SliceStable(g.errorChecks, func(i, j int) bool {
		a, b := g.errorChecks[i], g.errorChecks[j]
		return a.Type < b.Type
	})
}
