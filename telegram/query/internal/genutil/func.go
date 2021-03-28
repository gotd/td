package genutil

import (
	"go/types"

	"golang.org/x/tools/go/packages"
)

// Func is a function representation.
type Func struct {
	Sig  *types.Signature
	Decl *types.Func
}

// Results returns function results.
func (f Func) Results() *types.Tuple {
	return f.Sig.Results()
}

// Args returns function arguments.
func (f Func) Args() *types.Tuple {
	return f.Sig.Params()
}

// Funcs collects all function from package using given filter.
// Parameter keep may be nil.
func Funcs(pkg *packages.Package, keep func(f Func) bool) []Func {
	var r []Func

	for _, def := range pkg.TypesInfo.Defs {
		if def == nil {
			continue
		}

		f, ok := def.(*types.Func)
		if !ok {
			continue
		}

		sig, ok := f.Type().(*types.Signature)
		if !ok {
			continue
		}
		repr := Func{
			Sig:  sig,
			Decl: f,
		}

		if keep(repr) {
			r = append(r, repr)
		}
	}

	return r
}
